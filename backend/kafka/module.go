package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/99designs/keyring"
	"github.com/adrg/xdg"
	"github.com/jinzhu/copier"
	"go.uber.org/multierr"
)

type State struct {
	Projects map[string]*StateProject `json:"projects"`
}

type StateProject struct {
	ID           string `json:"id"`
	CurrentTab   string `json:"currentTab"`
	Address      string `json:"address"`
	AuthMethod   string `json:"authMethod"`
	AuthUsername string `json:"authUsername"`
	AuthPassword string `json:"authPassword" copier:"-"`
}

type Module struct {
	AppCtx         context.Context
	configFilePath string
	ring           keyring.Keyring
	state          *State
	stateMutex     *sync.RWMutex
}

func NewModule() (*Module, error) {
	configFilePath, err := xdg.ConfigFile("multibase/kafka.json")
	if err != nil {
		return nil, fmt.Errorf("failed to resolve kafka config path: %w", err)
	}

	ring, err := keyring.Open(keyring.Config{
		ServiceName:              "multibase_kafka",
		KeychainTrustApplication: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open kafka keyring: %w", err)
	}

	module := &Module{
		configFilePath: configFilePath,
		ring:           ring,
		state: &State{
			Projects: make(map[string]*StateProject),
		},
		stateMutex: &sync.RWMutex{},
	}

	err = module.readOrInitializeState()
	if err != nil {
		return nil, err
	}

	return module, nil
}

func (m *Module) CreateNewProject(projectID string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.state.Projects[projectID] = &StateProject{
		ID:         projectID,
		CurrentTab: "overview",
		Address:    "0.0.0.0:9092",
		AuthMethod: "plaintext",
	}

	if err := m.saveState(); err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) DeleteProject(projectID string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	delete(m.state.Projects, projectID)

	passwordKey := fmt.Sprintf("%s_AuthPassword", projectID)
	_ = m.ring.Remove(passwordKey)

	if err := m.saveState(); err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) SaveCurrentTab(projectID, currentTab string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.state.Projects[projectID].CurrentTab = currentTab

	if err := m.saveState(); err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) SaveAddress(projectID, address string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.state.Projects[projectID].Address = address

	if err := m.saveState(); err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) SaveAuthMethod(projectID, authMethod string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.state.Projects[projectID].AuthMethod = authMethod

	if authMethod == "plaintext" {
		passwordKey := fmt.Sprintf("%s_AuthPassword", projectID)
		_ = m.ring.Remove(passwordKey)

		m.state.Projects[projectID].AuthUsername = ""
		m.state.Projects[projectID].AuthPassword = ""
	}

	if err := m.saveState(); err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) SaveAuthUsername(projectID, authUsername string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.state.Projects[projectID].AuthUsername = authUsername

	if err := m.saveState(); err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) SaveAuthPassword(projectID, authPassword string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.state.Projects[projectID].AuthPassword = authPassword

	passwordKey := fmt.Sprintf("%s_AuthPassword", projectID)

	if authPassword == "" {
		_ = m.ring.Remove(passwordKey)
	} else {
		err := m.ring.Set(keyring.Item{
			Key:  passwordKey,
			Data: []byte(authPassword),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to set a kafka keyring key: %w", err)
		}
	}

	if err := m.saveState(); err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) State() (*State, error) {
	m.stateMutex.RLock()
	defer m.stateMutex.RUnlock()

	return m.state, nil
}

func (m *Module) readOrInitializeState() error {
	_, err := os.Stat(m.configFilePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to describe a kafka config file: %w", err)
		}

		return m.initializeState()
	}

	return m.readState()
}

func (m *Module) initializeState() (rerr error) {
	file, err := os.Create(m.configFilePath)
	if err != nil {
		return fmt.Errorf("failed to create a kafka config file: %w", err)
	}

	defer func() {
		err := file.Close()
		if err != nil {
			rerr = multierr.Combine(rerr, fmt.Errorf("failed to close a config file: %w", err))
		}
	}()

	encoder := json.NewEncoder(file)

	err = encoder.Encode(m.state)
	if err != nil {
		return fmt.Errorf("failed to encode a kafka state: %w", err)
	}

	return nil
}

func (m *Module) readState() (rerr error) {
	file, err := os.Open(m.configFilePath)
	if err != nil {
		return fmt.Errorf("failed to open a kafka config file: %w", err)
	}

	defer func() {
		err := file.Close()
		if err != nil {
			rerr = multierr.Combine(rerr, fmt.Errorf("failed to close a config file: %w", err))
		}
	}()

	decoder := json.NewDecoder(file)

	err = decoder.Decode(&m.state)
	if err != nil {
		return fmt.Errorf("failed to decode a kafka state: %w", err)
	}

	for _, project := range m.state.Projects {
		item, err := m.ring.Get(fmt.Sprintf("%s_AuthPassword", project.ID))
		if err != nil {
			if errors.Is(err, keyring.ErrKeyNotFound) {
				continue
			}

			return fmt.Errorf("failed to set a kafka keyring key: %w", err)
		}

		project.AuthPassword = string(item.Data)
	}

	return nil
}

func (m *Module) saveState() (rerr error) {
	file, err := os.Create(m.configFilePath)
	if err != nil {
		return fmt.Errorf("failed to create/truncate a kafka config file: %w", err)
	}

	defer func() {
		err := file.Close()
		if err != nil {
			rerr = multierr.Combine(rerr, fmt.Errorf("failed to close a config file: %w", err))
		}
	}()

	state := &State{}

	err = copier.CopyWithOption(state, m.state, copier.Option{IgnoreEmpty: true, DeepCopy: true})
	if err != nil {
		return fmt.Errorf("failed to copy a kafka state: %w", err)
	}

	encoder := json.NewEncoder(file)

	err = encoder.Encode(state)
	if err != nil {
		return fmt.Errorf("failed to encode a kafka state: %w", err)
	}

	return nil
}

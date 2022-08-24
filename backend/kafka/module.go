package kafka

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/99designs/keyring"
	"github.com/adrg/xdg"
	"github.com/jinzhu/copier"
	"go.uber.org/multierr"
	"golang.org/x/crypto/scrypt"
)

var defaultPassword = []byte("###multibase_storage_password###") // nolint: gochecknoglobals

const (
	defaultStatePersistenceDelay = time.Second * 3
)

type State struct {
	Projects map[string]*StateProject `json:"projects"`
}

type StateProject struct {
	ID           string `json:"id"`
	CurrentTab   string `json:"currentTab"`
	Address      string `json:"address"`
	AuthMethod   string `json:"authMethod"`
	AuthUsername string `json:"authUsername" copier:"-"`
	AuthPassword string `json:"authPassword" copier:"-"`
}

type Module struct {
	AppCtx         context.Context
	configFilePath string
	ring           keyring.Keyring
	state          *State
	stateMutex     *sync.RWMutex
	stateTimer     *time.Timer
}

func NewModule() (*Module, error) {
	configFilePath, err := xdg.ConfigFile("multibase/kafka")
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

func (m *Module) SaveState() error {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	if m.stateTimer != nil {
		_ = m.stateTimer.Stop()
	}

	if err := m.saveStateToFile(); err != nil {
		return fmt.Errorf("failed to save state to file: %w", err)
	}

	return nil
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

	if err := m.saveStateToFile(); err != nil {
		return nil, fmt.Errorf("failed to save state to file: %w", err)
	}

	return m.state, nil
}

func (m *Module) DeleteProject(projectID string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	delete(m.state.Projects, projectID)

	passwordKey := fmt.Sprintf("%s_AuthPassword", projectID)
	_ = m.ring.Remove(passwordKey)

	m.saveState()

	return m.state, nil
}

func (m *Module) SaveCurrentTab(projectID, currentTab string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.state.Projects[projectID].CurrentTab = currentTab

	m.saveState()

	return m.state, nil
}

func (m *Module) SaveAddress(projectID, address string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.state.Projects[projectID].Address = address

	m.saveState()

	return m.state, nil
}

func (m *Module) SaveAuthMethod(projectID, authMethod string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.state.Projects[projectID].AuthMethod = authMethod

	if authMethod == "plaintext" {
		passwordKey := fmt.Sprintf("%s_AuthPassword", projectID)
		usernameKey := fmt.Sprintf("%s_AuthUsername", projectID)
		_ = m.ring.Remove(passwordKey)
		_ = m.ring.Remove(usernameKey)

		m.state.Projects[projectID].AuthUsername = ""
		m.state.Projects[projectID].AuthPassword = ""
	}

	m.saveState()

	return m.state, nil
}

func (m *Module) SaveAuthUsername(projectID, authUsername string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.state.Projects[projectID].AuthUsername = authUsername

	usernameKey := fmt.Sprintf("%s_AuthUsername", projectID)

	if authUsername == "" {
		_ = m.ring.Remove(usernameKey)
	} else {
		err := m.ring.Set(keyring.Item{
			Key:  usernameKey,
			Data: []byte(authUsername),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to set a kafka keyring key: %w", err)
		}
	}

	m.saveState()

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

	m.saveState()

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

	data, err := json.Marshal(m.state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	encryptedState, err := encrypt(defaultPassword, data)
	if err != nil {
		return fmt.Errorf("failed to encrypt state: %w", err)
	}

	_, err = file.Write(encryptedState)
	if err != nil {
		return fmt.Errorf("failed to write state: %w", err)
	}

	return nil
}

// nolint: cyclop
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

	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read state from file: %w", err)
	}

	decryptedData, err := decrypt(defaultPassword, data)
	if err != nil {
		return fmt.Errorf("failed to decrypt state: %w", err)
	}

	err = json.Unmarshal(decryptedData, m.state)
	if err != nil {
		return fmt.Errorf("failed to unmarshal state: %w", err)
	}

	for _, project := range m.state.Projects {
		if project.AuthMethod == "plaintext" {
			continue
		}

		authPassword, err := m.ring.Get(fmt.Sprintf("%s_AuthPassword", project.ID))
		if err != nil && !errors.Is(err, keyring.ErrKeyNotFound) {
			return fmt.Errorf("failed to get a kafka keyring key: %w", err)
		}

		project.AuthPassword = string(authPassword.Data)

		authUsername, err := m.ring.Get(fmt.Sprintf("%s_AuthUsername", project.ID))
		if err != nil && !errors.Is(err, keyring.ErrKeyNotFound) {
			return fmt.Errorf("failed to get a kafka keyring key: %w", err)
		}

		project.AuthUsername = string(authUsername.Data)
	}

	return nil
}

func (m *Module) saveState() {
	if m.stateTimer != nil {
		_ = m.stateTimer.Stop()
	}

	m.stateTimer = time.AfterFunc(defaultStatePersistenceDelay, func() {
		err := m.saveStateToFile()
		if err != nil {
			log.Println(fmt.Errorf("failed to save state to a file: %w", err))
		}
	})
}

func (m *Module) saveStateToFile() (rerr error) {
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

	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	encryptedData, err := encrypt(defaultPassword, data)
	if err != nil {
		return fmt.Errorf("failed to encrypt state: %w", err)
	}

	_, err = file.Write(encryptedData)
	if err != nil {
		return fmt.Errorf("failed to write state: %w", err)
	}

	return nil
}

func encrypt(key, data []byte) ([]byte, error) {
	key, salt, err := deriveKey(key, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to derive key: %w", err)
	}

	blockCipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(blockCipher)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new gcm: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to prepare nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)

	ciphertext = append(ciphertext, salt...)

	return ciphertext, nil
}

func decrypt(key, data []byte) ([]byte, error) {
	salt, data := data[len(data)-32:], data[:len(data)-32]

	key, _, err := deriveKey(key, salt)
	if err != nil {
		return nil, fmt.Errorf("failed to derive key: %w", err)
	}

	blockCipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(blockCipher)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new gcm: %w", err)
	}

	nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open gcm: %w", err)
	}

	return plaintext, nil
}

func deriveKey(password, salt []byte) ([]byte, []byte, error) {
	if salt == nil {
		const defaultSaltLen = 32

		salt = make([]byte, defaultSaltLen)
		if _, err := rand.Read(salt); err != nil {
			return nil, nil, fmt.Errorf("failed to prepate salt: %w", err)
		}
	}

	// nolint: varnamelen
	const (
		n      = 4096
		r      = 8
		p      = 1
		keyLen = 32
	)

	key, err := scrypt.Key(password, salt, n, r, p, keyLen)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate scrypt key: %w", err)
	}

	return key, salt, nil
}

package project

import (
	"fmt"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/samber/lo"

	"github.com/multibase-io/multibase/backend/pkg/state"
)

type State struct {
	Stats            *StateStats              `json:"stats"`
	Projects         map[string]*StateProject `json:"projects"`
	OpenedProjectIDs []string                 `json:"openedProjectIDs"`
	CurrentProjectID string                   `json:"currentProjectID"`
}

type StateProject struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
}

type StateStats struct {
	GRPCProjectCount   int `json:"grpcProjectCount"`
	ThriftProjectCount int `json:"thriftProjectCount"`
	KafkaProjectCount  int `json:"kafkaProjectCount"`
}

type Module struct {
	stateStorage *state.Storage
	stateMutex   *sync.RWMutex
	state        *State
}

func NewModule(stateStorage *state.Storage) (*Module, error) {
	module := &Module{
		state: &State{
			Projects: make(map[string]*StateProject),
		},
		stateStorage: stateStorage,
		stateMutex:   &sync.RWMutex{},
	}

	err := module.readOrInitializeState()
	if err != nil {
		return nil, err
	}

	return module, nil
}

func (m *Module) OpenProject(newProjectID, projectToOpenID string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	if lo.Contains(m.state.OpenedProjectIDs, projectToOpenID) {
		m.state.OpenedProjectIDs = lo.Reject(m.state.OpenedProjectIDs, func(projectID string, _ int) bool {
			return projectID == newProjectID
		})
	} else {
		m.state.OpenedProjectIDs = lo.ReplaceAll(m.state.OpenedProjectIDs, newProjectID, projectToOpenID)
	}

	m.state.CurrentProjectID = projectToOpenID
	delete(m.state.Projects, newProjectID)

	if err := m.saveState(); err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) CreateGRPCProject(projectID string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.state.Stats.GRPCProjectCount++

	projectName := fmt.Sprintf("gRPC Project %d", m.state.Stats.GRPCProjectCount)

	m.state.Projects[projectID] = &StateProject{
		ID:   projectID,
		Type: "grpc",
		Name: projectName,
	}

	if err := m.saveState(); err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) CreateThriftProject(projectID string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.state.Stats.ThriftProjectCount++

	projectName := fmt.Sprintf("Thrift Project %d", m.state.Stats.ThriftProjectCount)

	m.state.Projects[projectID] = &StateProject{
		ID:   projectID,
		Type: "thrift",
		Name: projectName,
	}

	if err := m.saveState(); err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) CreateKafkaProject(projectID string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.state.Stats.KafkaProjectCount++

	projectName := fmt.Sprintf("Kafka Project %d", m.state.Stats.KafkaProjectCount)

	m.state.Projects[projectID] = &StateProject{
		ID:   projectID,
		Type: "kafka",
		Name: projectName,
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

	m.state.OpenedProjectIDs = lo.Reject(m.state.OpenedProjectIDs, func(pID string, _ int) bool {
		return pID == projectID
	})

	if m.state.CurrentProjectID == projectID {
		m.state.CurrentProjectID = m.state.OpenedProjectIDs[len(m.state.OpenedProjectIDs)-1]
	}

	if err := m.saveState(); err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) RenameProject(projectID, name string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.state.Projects[projectID].Name = name

	if err := m.saveState(); err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) CreateNewProject() (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	projectID := uuid.Must(uuid.NewV4()).String()

	m.state.Projects[projectID] = &StateProject{
		ID:   projectID,
		Type: "new",
	}

	m.state.OpenedProjectIDs = append(m.state.OpenedProjectIDs, projectID)
	m.state.CurrentProjectID = projectID

	if err := m.saveState(); err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) CloseProject(projectID string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	if len(m.state.OpenedProjectIDs) <= 1 {
		return m.state, nil
	}

	if m.state.Projects[projectID].Type == "new" {
		delete(m.state.Projects, projectID)
	}

	m.state.OpenedProjectIDs = lo.Reject(m.state.OpenedProjectIDs, func(pID string, _ int) bool {
		return pID == projectID
	})

	if m.state.CurrentProjectID == projectID {
		m.state.CurrentProjectID = m.state.OpenedProjectIDs[len(m.state.OpenedProjectIDs)-1]
	}

	if err := m.saveState(); err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) SaveCurrentProjectID(projectID string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.state.CurrentProjectID = projectID

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
	isLoaded, err := m.stateStorage.Load("project", m.state)
	if err != nil {
		return fmt.Errorf("failed to load a state: %w", err)
	}

	if isLoaded {
		return nil
	}

	m.state = &State{
		Stats: &StateStats{
			GRPCProjectCount:   0,
			ThriftProjectCount: 0,
			KafkaProjectCount:  0,
		},
		Projects: map[string]*StateProject{
			"404f5702-6179-4861-9533-b5ee16161c78": {
				ID:   "404f5702-6179-4861-9533-b5ee16161c78",
				Type: "new",
			},
		},
		OpenedProjectIDs: []string{"404f5702-6179-4861-9533-b5ee16161c78"},
		CurrentProjectID: "404f5702-6179-4861-9533-b5ee16161c78",
	}

	err = m.stateStorage.Save("project", m.state)
	if err != nil {
		return fmt.Errorf("failed to store a state: %w", err)
	}

	return nil
}

func (m *Module) saveState() error {
	err := m.stateStorage.Save("project", m.state)
	if err != nil {
		return fmt.Errorf("failed to store a state: %w", err)
	}

	return nil
}

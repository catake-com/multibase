package project

import (
	"fmt"
	"sync"

	"github.com/gofrs/uuid/v5"
	"github.com/samber/lo"

	"github.com/multibase-io/multibase/backend/pkg/state"
)

type Module struct {
	Stats            *Stats              `json:"stats"`
	Projects         map[string]*Project `json:"projects"`
	OpenedProjectIDs []string            `json:"openedProjectIDs"`
	CurrentProjectID string              `json:"currentProjectID"`

	stateStorage *state.Storage
	stateMutex   *sync.RWMutex
}

func NewModule(stateStorage *state.Storage) (*Module, error) {
	module := &Module{
		Projects:     make(map[string]*Project),
		stateStorage: stateStorage,
		stateMutex:   &sync.RWMutex{},
	}

	err := module.readOrInitializeState()
	if err != nil {
		return nil, err
	}

	return module, nil
}

func (m *Module) OpenProject(newProjectID, projectToOpenID string) (*Module, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	if lo.Contains(m.OpenedProjectIDs, projectToOpenID) {
		m.OpenedProjectIDs = lo.Reject(m.OpenedProjectIDs, func(projectID string, _ int) bool {
			return projectID == newProjectID
		})
	} else {
		m.OpenedProjectIDs = lo.ReplaceAll(m.OpenedProjectIDs, newProjectID, projectToOpenID)
	}

	m.CurrentProjectID = projectToOpenID
	delete(m.Projects, newProjectID)

	if err := m.saveState(); err != nil {
		return nil, err
	}

	return m, nil
}

func (m *Module) CreateGRPCProject(projectID string) (*Module, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.Stats.GRPCProjectCount++

	projectName := fmt.Sprintf("gRPC Project %d", m.Stats.GRPCProjectCount)

	m.Projects[projectID] = &Project{
		ID:   projectID,
		Type: "grpc",
		Name: projectName,
	}

	if err := m.saveState(); err != nil {
		return nil, err
	}

	return m, nil
}

func (m *Module) CreateThriftProject(projectID string) (*Module, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.Stats.ThriftProjectCount++

	projectName := fmt.Sprintf("Thrift Project %d", m.Stats.ThriftProjectCount)

	m.Projects[projectID] = &Project{
		ID:   projectID,
		Type: "thrift",
		Name: projectName,
	}

	if err := m.saveState(); err != nil {
		return nil, err
	}

	return m, nil
}

func (m *Module) CreateKafkaProject(projectID string) (*Module, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.Stats.KafkaProjectCount++

	projectName := fmt.Sprintf("Kafka Project %d", m.Stats.KafkaProjectCount)

	m.Projects[projectID] = &Project{
		ID:   projectID,
		Type: "kafka",
		Name: projectName,
	}

	if err := m.saveState(); err != nil {
		return nil, err
	}

	return m, nil
}

func (m *Module) DeleteProject(projectID string) (*Module, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	delete(m.Projects, projectID)

	m.OpenedProjectIDs = lo.Reject(m.OpenedProjectIDs, func(pID string, _ int) bool {
		return pID == projectID
	})

	if m.CurrentProjectID == projectID {
		m.CurrentProjectID = m.OpenedProjectIDs[len(m.OpenedProjectIDs)-1]
	}

	if err := m.saveState(); err != nil {
		return nil, err
	}

	return m, nil
}

func (m *Module) RenameProject(projectID, name string) (*Module, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.Projects[projectID].Name = name

	if err := m.saveState(); err != nil {
		return nil, err
	}

	return m, nil
}

func (m *Module) CreateNewProject() (*Module, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	projectID := uuid.Must(uuid.NewV4()).String()

	m.Projects[projectID] = &Project{
		ID:   projectID,
		Type: "new",
	}

	m.OpenedProjectIDs = append(m.OpenedProjectIDs, projectID)
	m.CurrentProjectID = projectID

	if err := m.saveState(); err != nil {
		return nil, err
	}

	return m, nil
}

func (m *Module) CloseProject(projectID string) (*Module, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	if len(m.OpenedProjectIDs) <= 1 {
		return m, nil
	}

	if m.Projects[projectID].Type == "new" {
		delete(m.Projects, projectID)
	}

	m.OpenedProjectIDs = lo.Reject(m.OpenedProjectIDs, func(pID string, _ int) bool {
		return pID == projectID
	})

	if m.CurrentProjectID == projectID {
		m.CurrentProjectID = m.OpenedProjectIDs[len(m.OpenedProjectIDs)-1]
	}

	if err := m.saveState(); err != nil {
		return nil, err
	}

	return m, nil
}

func (m *Module) SaveCurrentProjectID(projectID string) (*Module, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.CurrentProjectID = projectID

	if err := m.saveState(); err != nil {
		return nil, err
	}

	return m, nil
}

func (m *Module) State() (*Module, error) {
	m.stateMutex.RLock()
	defer m.stateMutex.RUnlock()

	return m, nil
}

func (m *Module) readOrInitializeState() error {
	isLoaded, err := m.stateStorage.Load("project", m)
	if err != nil {
		return fmt.Errorf("failed to load a state: %w", err)
	}

	if isLoaded {
		return nil
	}

	m.Stats = &Stats{
		GRPCProjectCount:   0,
		ThriftProjectCount: 0,
		KafkaProjectCount:  0,
	}
	m.Projects["404f5702-6179-4861-9533-b5ee16161c78"] = &Project{
		ID:   "404f5702-6179-4861-9533-b5ee16161c78",
		Type: "new",
	}
	m.OpenedProjectIDs = []string{"404f5702-6179-4861-9533-b5ee16161c78"}
	m.CurrentProjectID = "404f5702-6179-4861-9533-b5ee16161c78"

	err = m.stateStorage.Save("project", m)
	if err != nil {
		return fmt.Errorf("failed to store a state: %w", err)
	}

	return nil
}

func (m *Module) saveState() error {
	err := m.stateStorage.Save("project", m)
	if err != nil {
		return fmt.Errorf("failed to store a state: %w", err)
	}

	return nil
}

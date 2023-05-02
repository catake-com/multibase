package kafka

import (
	"context"
	"fmt"
	"sync"

	"github.com/multibase-io/multibase/backend/pkg/state"
)

type Module struct {
	AppCtx context.Context

	projects      map[string]*Project
	projectsMutex sync.RWMutex
	stateStorage  *state.Storage
}

func NewModule(stateStorage *state.Storage) (*Module, error) {
	module := &Module{
		projects:     make(map[string]*Project),
		stateStorage: stateStorage,
	}

	return module, nil
}

func (m *Module) CreateNewProject(projectID string) (*State, error) {
	project, err := NewProject(projectID, m.stateStorage)
	if err != nil {
		return nil, err
	}

	return project.state, nil
}

func (m *Module) DeleteProject(projectID string) error {
	project, err := m.fetchProject(projectID)
	if err != nil {
		return err
	}

	m.projectsMutex.Lock()
	defer m.projectsMutex.Unlock()

	err = project.Close()
	if err != nil {
		return err
	}

	err = m.stateStorage.Delete(projectID)
	if err != nil {
		return fmt.Errorf("failed to delete a state: %w", err)
	}

	delete(m.projects, projectID)

	return nil
}

func (m *Module) SaveState(projectID string, state *State) (*State, error) {
	project, err := m.fetchProject(projectID)
	if err != nil {
		return nil, err
	}

	err = project.SaveState(state)
	if err != nil {
		return nil, err
	}

	return project.state, nil
}

func (m *Module) Connect(projectID string) (*State, error) {
	project, err := m.fetchProject(projectID)
	if err != nil {
		return nil, err
	}

	err = project.Connect()
	if err != nil {
		return nil, err
	}

	return project.state, nil
}

func (m *Module) Topics(projectID string) (*TabTopicsData, error) {
	project, err := m.fetchProject(projectID)
	if err != nil {
		return nil, err
	}

	data, err := project.Topics()
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (m *Module) Brokers(projectID string) (*TabBrokersData, error) {
	project, err := m.fetchProject(projectID)
	if err != nil {
		return nil, err
	}

	data, err := project.Brokers()
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (m *Module) Consumers(projectID string) (*TabConsumersData, error) {
	project, err := m.fetchProject(projectID)
	if err != nil {
		return nil, err
	}

	data, err := project.Consumers()
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (m *Module) StartTopicConsuming(
	projectID string,
	consumingStrategy TopicConsumingStrategy,
	topic,
	timeFrom string,
	offsetValue int64,
) (*TopicOutput, error) {
	project, err := m.fetchProject(projectID)
	if err != nil {
		return nil, err
	}

	data, err := project.StartTopicConsuming(m.AppCtx, consumingStrategy, topic, timeFrom, offsetValue)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (m *Module) StopTopicConsuming(projectID string) error {
	project, err := m.fetchProject(projectID)
	if err != nil {
		return err
	}

	err = project.StopTopicConsuming()
	if err != nil {
		return err
	}

	return nil
}

func (m *Module) ProjectState(projectID string) (*State, error) {
	project, err := m.fetchProject(projectID)
	if err != nil {
		return nil, err
	}

	return project.state, nil
}

func (m *Module) Close() error {
	for _, project := range m.projects {
		err := project.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Module) fetchProject(projectID string) (*Project, error) {
	m.projectsMutex.RLock()
	project, ok := m.projects[projectID]
	m.projectsMutex.RUnlock()

	if ok {
		return project, nil
	}

	project = &Project{}
	projectState := &State{}

	isLoaded, err := m.stateStorage.Load(projectID, projectState)
	if err != nil {
		return nil, fmt.Errorf("failed to load a state: %w", err)
	}

	if !isLoaded {
		return nil, nil
	}

	projectState.CurrentTab = TabOverview
	project.state = projectState
	project.stateStorage = m.stateStorage

	m.projectsMutex.Lock()
	m.projects[projectID] = project
	m.projectsMutex.Unlock()

	return project, nil
}

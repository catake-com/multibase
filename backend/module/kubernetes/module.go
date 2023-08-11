package kubernetes

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/catake-com/multibase/backend/state"
)

type Module struct {
	AppCtx context.Context

	projects      map[string]*Project
	projectsMutex sync.RWMutex
	stateStorage  *state.Storage
	appLogger     *logrus.Logger
}

func NewModule(stateStorage *state.Storage, appLogger *logrus.Logger) (*Module, error) {
	module := &Module{
		projects:     make(map[string]*Project),
		stateStorage: stateStorage,
		appLogger:    appLogger,
	}

	module.extendEnvPath()

	return module, nil
}

func (m *Module) CreateNewProject(projectID string) (*State, error) {
	project, err := NewProject(projectID, m.stateStorage, m.appLogger)
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

func (m *Module) SaveCurrentTab(projectID string, currentTab Tab) (*State, error) {
	project, err := m.fetchProject(projectID)
	if err != nil {
		return nil, err
	}

	err = project.SaveCurrentTab(currentTab)
	if err != nil {
		return nil, err
	}

	return project.state, nil
}

func (m *Module) SelectNamespace(projectID, selectedNamespace string) (*State, error) {
	project, err := m.fetchProject(projectID)
	if err != nil {
		return nil, err
	}

	err = project.SelectNamespace(selectedNamespace)
	if err != nil {
		return nil, err
	}

	return project.state, nil
}

func (m *Module) Connect(projectID, selectedContext string) (*State, error) {
	project, err := m.fetchProject(projectID)
	if err != nil {
		return nil, err
	}

	err = project.Connect(selectedContext)
	if err != nil {
		return nil, err
	}

	return project.state, nil
}

func (m *Module) StartPortForwarding(projectID, namespace, pod, ports string) (*State, error) {
	project, err := m.fetchProject(projectID)
	if err != nil {
		return nil, err
	}

	err = project.StartPortForwarding(namespace, pod, ports)
	if err != nil {
		return nil, err
	}

	return project.state, nil
}

func (m *Module) StopPortForwarding(projectID string) (*State, error) {
	project, err := m.fetchProject(projectID)
	if err != nil {
		return nil, err
	}

	err = project.StopPortForwarding()
	if err != nil {
		return nil, err
	}

	return project.state, nil
}

func (m *Module) Namespaces(projectID string) ([]string, error) {
	project, err := m.fetchProject(projectID)
	if err != nil {
		return nil, err
	}

	return project.Namespaces()
}

func (m *Module) OverviewData(projectID string) (*TabOverviewData, error) {
	project, err := m.fetchProject(projectID)
	if err != nil {
		return nil, err
	}

	data, err := project.OverviewData()
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (m *Module) WorkloadsPodsData(projectID string) (*TabWorkloadsPodsData, error) {
	project, err := m.fetchProject(projectID)
	if err != nil {
		return nil, err
	}

	return project.WorkloadsPodsData()
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
	project.appLogger = m.appLogger

	err = project.Initialize()
	if err != nil {
		return nil, err
	}

	m.projectsMutex.Lock()
	m.projects[projectID] = project
	m.projectsMutex.Unlock()

	return project, nil
}

func (m *Module) extendEnvPath() {
	shellName := getShellName()
	configFilePath := getShellConfigFilePath(shellName)

	if shellName == "" || configFilePath == "" {
		m.appLogger.Info("no shell or shell config file detected")

		return
	}

	cmd := exec.Command(shellName, "-c", fmt.Sprintf("source %s && env", configFilePath))

	output, err := cmd.Output()
	if err != nil {
		m.appLogger.Error(fmt.Errorf("failed to exec env: %w", err))

		return
	}

	envVars := parseEnvOutput(string(output))

	if envVars["PATH"] != "" {
		err := os.Setenv("PATH", fmt.Sprintf("%s:%s", os.Getenv("PATH"), envVars["PATH"]))
		if err != nil {
			m.appLogger.Error(fmt.Errorf("failed to set env: %w", err))

			return
		}
	}
}

func getShellName() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		return ""
	}

	return filepath.Base(shell)
}

func getShellConfigFilePath(shell string) string {
	switch shell {
	case "bash":
		for _, cfgName := range []string{".bashrc", ".bash_profile"} {
			path := filepath.Join(os.Getenv("HOME"), cfgName)

			_, err := os.Stat(path)
			if err != nil {
				continue
			}

			return path
		}
	case "zsh":
		for _, cfgName := range []string{".zshrc", ".zprofile"} {
			path := filepath.Join(os.Getenv("HOME"), cfgName)

			_, err := os.Stat(path)
			if err != nil {
				continue
			}

			return path
		}
	}

	return ""
}

// nolint: gomnd
func parseEnvOutput(output string) map[string]string {
	envVars := make(map[string]string)

	lines := strings.Split(output, "\n")

	for _, line := range lines {
		parts := strings.SplitN(line, "=", 2)

		if len(parts) == 2 {
			envVars[parts[0]] = parts[1]
		}
	}

	return envVars
}

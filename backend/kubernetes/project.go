package kubernetes

import (
	"fmt"
	"os"
	"path"
	"sort"
	"sync"

	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"

	"github.com/multibase-io/multibase/backend/pkg/state"
)

type Project struct {
	state               *State
	stateMutex          sync.RWMutex
	stateStorage        *state.Storage
	appLogger           *logrus.Logger
	apiConfig           api.Config
	restConfig          *rest.Config
	kubernetesClientset *kubernetes.Clientset
}

func NewProject(projectID string, stateStorage *state.Storage, appLogger *logrus.Logger) (*Project, error) {
	project := &Project{
		state: &State{
			ID:         projectID,
			CurrentTab: TabOverview,
		},
		stateStorage: stateStorage,
		appLogger:    appLogger,
	}

	err := project.Initialize()
	if err != nil {
		return nil, err
	}

	if err := project.saveState(); err != nil {
		return nil, err
	}

	return project, nil
}

func (p *Project) Initialize() error {
	homeDirPath, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to fetch home dir path: %w", err)
	}

	apiConfig, err := clientcmd.
		NewNonInteractiveDeferredLoadingClientConfig(
			&clientcmd.ClientConfigLoadingRules{ExplicitPath: path.Join(homeDirPath, ".kube/config")},
			&clientcmd.ConfigOverrides{},
		).
		RawConfig()
	if err != nil {
		return fmt.Errorf("failed to build api config: %w", err)
	}

	restConfig, err := clientcmd.BuildConfigFromFlags("", path.Join(homeDirPath, ".kube/config"))
	if err != nil {
		return fmt.Errorf("failed to build rest config: %w", err)
	}

	p.apiConfig = apiConfig
	p.restConfig = restConfig

	return nil
}

func (p *Project) SaveCurrentTab(currentTab Tab) error {
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()

	p.state.CurrentTab = currentTab

	return p.saveState()
}

func (p *Project) SelectNamespace(selectedNamespace string) error {
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()

	p.state.SelectedNamespace = selectedNamespace

	return p.saveState()
}

func (p *Project) Connect(selectedContext string) error {
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()

	clientset, err := kubernetes.NewForConfig(p.restConfig)
	if err != nil {
		return fmt.Errorf("failed to build clientset: %w", err)
	}

	p.state.IsConnected = true
	p.state.SelectedContext = selectedContext
	p.kubernetesClientset = clientset

	return p.saveState()
}

func (p *Project) OverviewData() (*TabOverviewData, error) {
	overviewData := &TabOverviewData{
		Contexts: make([]*TabOverviewDataContext, 0, len(p.apiConfig.Contexts)),
	}

	for name, context := range p.apiConfig.Contexts {
		overviewData.Contexts = append(
			overviewData.Contexts,
			&TabOverviewDataContext{
				IsSelected: name == p.state.SelectedContext,
				Name:       name,
				Cluster:    context.Cluster,
			},
		)
	}

	sort.Slice(overviewData.Contexts, func(i, j int) bool {
		return overviewData.Contexts[i].Name < overviewData.Contexts[j].Name
	})

	return overviewData, nil
}

func (p *Project) Close() error {
	return nil
}

func (p *Project) saveState() error {
	copiedState := *p.state
	copiedState.IsConnected = false
	copiedState.CurrentTab = ""

	err := p.stateStorage.Save(p.state.ID, &copiedState)
	if err != nil {
		return fmt.Errorf("failed to store a kafka project: %w", err)
	}

	return nil
}

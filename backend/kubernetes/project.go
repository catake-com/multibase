package kubernetes

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"sort"
	"sync"
	"time"

	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"

	"github.com/catake-com/multibase/backend/pkg/state"
)

const (
	requestTimeout = 5 * time.Second
)

var (
	errPortsAlreadyForwarded = errors.New("ports already forwarded")
	errPortsAreNotForwarded  = errors.New("ports are not forwarded")
)

type Project struct {
	state                  *State
	stateMutex             sync.RWMutex
	stateStorage           *state.Storage
	appLogger              *logrus.Logger
	apiConfig              api.Config
	restConfig             *rest.Config
	kubernetesClientset    *kubernetes.Clientset
	portForwardingStopChan chan struct{}
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

	restConfig.Timeout = requestTimeout

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

	_, err = p.kubernetesClientset.ServerVersion()
	if err != nil {
		return fmt.Errorf("failed to connect to kubernetes: %w", err)
	}

	return p.saveState()
}

func (p *Project) StartPortForwarding(namespace, pod, ports string) error {
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()

	if p.portForwardingStopChan != nil {
		return errPortsAlreadyForwarded
	}

	requestURL := p.kubernetesClientset.
		CoreV1().
		RESTClient().
		Post().
		Resource("pods").
		Namespace(namespace).
		Name(pod).
		SubResource("portforward").
		URL()

	transport, upgrader, err := spdy.RoundTripperFor(p.restConfig)
	if err != nil {
		return fmt.Errorf("failed to init round tripper: %w", err)
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, requestURL)

	p.portForwardingStopChan = make(chan struct{})
	p.state.IsPortForwarded = true

	portForwarder, err := portforward.New(
		dialer,
		[]string{ports},
		p.portForwardingStopChan,
		nil,
		os.Stdout,
		os.Stdout,
	)
	if err != nil {
		return fmt.Errorf("failed to init port forwarder: %w", err)
	}

	go func() {
		err := portForwarder.ForwardPorts()
		if err != nil {
			p.appLogger.Error(fmt.Errorf("failed to forward ports: %w", err))
		}
	}()

	return p.saveState()
}

func (p *Project) StopPortForwarding() error {
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()

	if p.portForwardingStopChan == nil {
		return errPortsAreNotForwarded
	}

	close(p.portForwardingStopChan)

	p.portForwardingStopChan = nil
	p.state.IsPortForwarded = false

	return p.saveState()
}

func (p *Project) Namespaces() ([]string, error) {
	ctx := context.Background()

	namespaces, err := p.kubernetesClientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		if err != nil {
			return nil, fmt.Errorf("failed to fetch namespaces: %w", err)
		}
	}

	namespaceNames := lo.Map(namespaces.Items, func(namespace v1.Namespace, _ int) string {
		return namespace.GetName()
	})

	sort.Slice(namespaceNames, func(i, j int) bool {
		return namespaceNames[i] < namespaceNames[j]
	})

	return namespaceNames, nil
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

func (p *Project) WorkloadsPodsData() (*TabWorkloadsPodsData, error) {
	ctx := context.Background()

	pods, err := p.kubernetesClientset.CoreV1().Pods(p.state.SelectedNamespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		if err != nil {
			return nil, fmt.Errorf("failed to fetch pods: %w", err)
		}
	}

	podsData := &TabWorkloadsPodsData{
		Pods: make([]*TabWorkloadsPodsDataPod, 0, len(pods.Items)),
	}

	for _, pod := range pods.Items {
		ports := make([]*TabWorkloadsPodsDataPodPort, 0, len(pod.Spec.Containers))

		for _, container := range pod.Spec.Containers {
			for _, port := range container.Ports {
				ports = append(
					ports,
					&TabWorkloadsPodsDataPodPort{
						Name:          port.Name,
						ContainerPort: int(port.ContainerPort),
					},
				)
			}
		}

		sort.Slice(ports, func(i, j int) bool {
			return ports[i].Name < ports[j].Name
		})

		podsData.Pods = append(
			podsData.Pods,
			&TabWorkloadsPodsDataPod{
				Name:      pod.GetName(),
				Namespace: pod.GetNamespace(),
				Ports:     ports,
			},
		)
	}

	sort.Slice(podsData.Pods, func(i, j int) bool {
		return podsData.Pods[i].Name < podsData.Pods[j].Name
	})

	return podsData, nil
}

func (p *Project) Close() error {
	if p.portForwardingStopChan != nil {
		close(p.portForwardingStopChan)
	}

	return nil
}

func (p *Project) saveState() error {
	copiedState := *p.state
	copiedState.IsConnected = false
	copiedState.IsPortForwarded = false
	copiedState.CurrentTab = ""

	err := p.stateStorage.Save(p.state.ID, &copiedState)
	if err != nil {
		return fmt.Errorf("failed to store a kafka project: %w", err)
	}

	return nil
}

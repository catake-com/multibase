package kubernetes

type Tab string

const (
	TabOverview  = "overview"
	TabWorkloads = "workloads"
)

type State struct {
	ID                string `json:"id"`
	SelectedContext   string `json:"selectedContext"`
	SelectedNamespace string `json:"selectedNamespace"`
	IsConnected       bool   `json:"isConnected"`
	CurrentTab        Tab    `json:"currentTab"`
}

type TabOverviewData struct {
	IsConnected bool                      `json:"isConnected"`
	Contexts    []*TabOverviewDataContext `json:"contexts"`
}

type TabOverviewDataContext struct {
	IsSelected bool   `json:"isSelected"`
	Name       string `json:"name"`
	Cluster    string `json:"cluster"`
}

type TabWorkloadsPodsData struct {
	Pods []*TabWorkloadsPodsDataPod `json:"pods"`
}

type TabWorkloadsPodsDataPod struct {
	Name      string                         `json:"name"`
	Namespace string                         `json:"namespace"`
	Ports     []*TabWorkloadsPodsDataPodPort `json:"ports"`
}

type TabWorkloadsPodsDataPodPort struct {
	Name          string `json:"name"`
	ContainerPort int    `json:"containerPort"`
}

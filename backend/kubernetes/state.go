package kubernetes

type Tab string

const (
	TabOverview      = "overview"
	TabWorkloads     = "workloads"
	TabWorkloadsPods = "workloads_pods"
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
	IsConnected bool                       `json:"isConnected"`
	Pods        []*TabWorkloadsPodsDataPod `json:"pods"`
}

type TabWorkloadsPodsDataPod struct {
	Name string `json:"name"`
}

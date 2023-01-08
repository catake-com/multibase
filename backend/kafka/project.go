package kafka

type Project struct {
	ID           string `json:"id"`
	Address      string `json:"address"`
	AuthMethod   string `json:"authMethod"`
	AuthUsername string `json:"authUsername"`
	AuthPassword string `json:"authPassword"`
	IsConnected  bool   `json:"isConnected" copier:"-"`
	CurrentTab   string `json:"currentTab" copier:"-"`
}

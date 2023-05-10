package types

type PluginInfo struct {
	Name        string     `json:"name"`
	Version     string     `json:"version"`
	Category    PluginType `json:"category"`
	Description string     `json:"description"`
	Repo        string     `json:"repo"`
	ImageTarget string     `json:"imageTarget"`
	Path        string     `json:"path"`
}

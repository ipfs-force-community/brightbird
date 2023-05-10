package types

type PluginType string

const (
	Deploy   PluginType = "Deployer"
	TestExec PluginType = "Exec"
)

type PluginInfo struct {
	Name        string     `json:"name"`
	Version     string     `json:"version"`
	PluginType  PluginType `json:"pluginType"`
	Description string     `json:"description"`
	Repo        string     `json:"repo"`
	ImageTarget string     `json:"imageTarget"`

	PluginParams
}

type PluginParams struct {
	Dependencies []DependencyProperty `json:"dependencies"`
	Properties   []Property           `json:"properties"`
}

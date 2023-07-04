package types

// PluginType  type of plugin
// swagger:alias
type PluginType string

const (
	// Deploy deploy conponet
	Deploy PluginType = "Deployer"
	// TestExec test case
	TestExec PluginType = "Exec"
)

type PluginInfo struct {
	Name        string     `json:"name"`
	Version     string     `json:"version"`
	PluginType  PluginType `json:"pluginType"`
	Description string     `json:"description"`
	Repo        string     `json:"repo"`
	ImageTarget string     `json:"imageTarget"`

	PluginParams `bson:",inline"`
}

type PluginParams struct {
	InputProperties  []Property `json:"inputProperties"`
	OutputProperties []Property `json:"outputProperties"`
}

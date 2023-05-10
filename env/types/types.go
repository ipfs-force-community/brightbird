package types

type BootstrapPeers []string
type PluginType string

const (
	Deploy   PluginType = "Deployer"
	TestExec PluginType = "Exec"
)

func PtrString(str string) *string {
	return &str
}

func GetString(str *string) string {
	if str == nil {
		return ""
	}
	return *str
}

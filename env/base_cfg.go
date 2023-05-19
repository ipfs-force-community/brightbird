package env

type BaseConfig struct {
	CodeVersion  string `json:"codeVersion"` //todo allow config as tag commit id brance
	InstanceName string `json:"-"`           //plugin instance name
}

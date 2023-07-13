package env

type BaseConfig struct {
	// CodeVersion deploy image commit
	CodeVersion string `jsonschema:"-" json:"codeVersion"` //
	// InstanceName plugin instance name
	InstanceName string `jsonschema:"-" json:"instanceName"`
}

func NewBaseConfig(codeVersion, instance string) BaseConfig {
	return BaseConfig{
		CodeVersion:  codeVersion,
		InstanceName: instance,
	}
}

package env

type BaseConfig struct {
	// CodeVersion deploy image commit
	CodeVersion string `json:"-"` //
	// InstanceName plugin instance name
	InstanceName string `json:"-"`
}

func NewBaseConfig(codeVersion, instance string) BaseConfig {
	return BaseConfig{
		CodeVersion:  codeVersion,
		InstanceName: instance,
	}
}

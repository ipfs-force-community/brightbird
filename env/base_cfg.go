package env

type BaseConfig struct {
	// CodeVersion deploy image commit
	CodeVersion string `ignore:"-" json:"codeVersion"` //
	// InstanceName plugin instance name
	InstanceName string `ignore:"-" json:"instanceName"`
}

func NewBaseConfig(codeVersion, instance string) BaseConfig {
	return BaseConfig{
		CodeVersion:  codeVersion,
		InstanceName: instance,
	}
}

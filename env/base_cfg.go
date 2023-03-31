package env

type BaseConfig struct {
	CodeVersion string `json:"codeVersion"` //todo allow config as tag commit id brance
	//use for annotate service name
	SvcMap map[string]string `json:"-"`
}

type BaseRenderParams struct {
	PrivateRegistry string
}

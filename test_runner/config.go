package main

type Config struct {
	NameSpace   string
	PluginStore string
	TmpPath     string

	BootstrapPeers []string
	Timeout        int
	Listen         string

	MongoURL string
	DBName   string

	Mysql string

	TaskId   string
	Registry string

	LogLevel         string
	CustomProperties map[string]interface{}
}

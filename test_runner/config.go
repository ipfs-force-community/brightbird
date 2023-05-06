package main

type Config struct {
	NameSpace   string
	PluginStore string
	TmpPath     string

	BootstrapPeers []string
	Timeout        int
	Listen         string

	MongoUrl string
	DbName   string

	Mysql string

	TaskId          string
	PrivateRegistry string
}

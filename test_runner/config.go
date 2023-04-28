package main

type Config struct {
	PluginStore string
	SharedDir   string

	BootstrapPeers []string
	Timeout        int
	Listen         string

	MongoUrl string
	DbName   string

	Mysql string

	TaskId          string
	PrivateRegistry string
}

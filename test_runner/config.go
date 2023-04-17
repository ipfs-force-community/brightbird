package main

type Config struct {
	PluginStore    string
	BootstrapPeers []string
	Timeout        int
	Listen         string

	MongoUrl string
	DbName   string

	Mysql           string
	SharedDir       string
	TaskId          string
	PrivateRegistry string
}

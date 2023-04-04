package main

type Config struct {
	PluginStore     string
	BootstrapPeers  []string
	Timeout         int
	Listen          string
	Mysql           string
	MongoUrl        string
	DbName          string
	TaskId          string
	PrivateRegistry string
}

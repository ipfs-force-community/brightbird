package main

type Config struct {
	PluginStore     string
	BootstrapPeers  []string
	Timeout         int
	Mysql           string
	MongoUrl        string
	DbName          string
	TaskId          string
	PrivateRegistry string
}

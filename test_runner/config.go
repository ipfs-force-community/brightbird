package main

type Config struct {
	PluginStore    string
	BootstrapPeers []string
	Timeout        int
	MongoUrl       string
	TaskId         string
}

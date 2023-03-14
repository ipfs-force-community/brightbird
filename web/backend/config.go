package main

type Config struct {
	PluginStore string
	MongoUrl    string
	Listen      string
}

func DefaultConfig() Config {
	return Config{
		MongoUrl: "mongodb://localhost:27017",
		Listen:   "127.0.0.1:12356",
	}
}

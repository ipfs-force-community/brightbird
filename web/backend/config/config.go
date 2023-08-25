package config

type Config struct {
	StaticRoot   string
	NameSpace    string
	PluginStore  string
	RunnerConfig string

	MongoURL string
	DBName   string
	Listen   string
	LogLevel string

	WebhookURL string
	Mysql      string

	BuildSpace string

	Proxy          string
	GitToken       string
	DockerRegistry []DockerRegistry

	BuildWorkers []BuildWorkerConfig

	CustomProperties map[string]string
}

type BuildWorkerConfig struct {
	BuildSpace string
}

type DockerRegistry struct {
	URL      string
	UserName string
	Password string
	Type     string
	Push     bool
}

func DefaultConfig() Config {
	return Config{
		NameSpace: "default",
		LogLevel:  "INFO",
		MongoURL:  "mongodb://localhost:27017",
		Listen:    "0.0.0.0:12356",
		DockerRegistry: []DockerRegistry{
			{
				URL: "https://registry.hub.docker.com",
			},
		},
	}
}

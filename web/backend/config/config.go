package config

type Config struct {
	StaticRoot   string
	NameSpace    string
	PluginStore  string
	LogPath      string
	RunnerConfig string

	BootstrapPeers []string
	MongoUrl       string
	DbName         string
	Listen         string

	WebhookUrl string
	Mysql      string

	BuildSpace string

	Proxy          string
	GitToken       string
	DockerRegistry []DockerRegistry

	BuildWorkers []BuildWorkerConfig
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
		StaticRoot: "dist",
		NameSpace:  "default",
		MongoUrl:   "mongodb://localhost:27017",
		Listen:     "0.0.0.0:12356",
		DockerRegistry: []DockerRegistry{
			{
				URL: "https://registry.hub.docker.com",
			},
		},
	}
}

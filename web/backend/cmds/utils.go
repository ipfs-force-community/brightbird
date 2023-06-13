package cmds

import "github.com/hunjixin/brightbird/web/backend/client"

func DefaulAPI() *client.BrightBirdAPI {
	return client.NewHTTPClientWithConfig(nil, &client.TransportConfig{
		Host:     "127.0.0.1:12356",
		BasePath: "/api/v1",
		Schemes:  []string{"http"},
	})
}

package main

import (
	"github.com/urfave/cli/v2"

	"github.com/ipfs-force-community/brightbird/internal/backend/client"
)

func DefaulAPI(cliCtx *cli.Context) *client.BrightBirdAPI {
	return client.NewHTTPClientWithConfig(nil, &client.TransportConfig{
		Host:     cliCtx.String("listen"),
		BasePath: "/api/v1",
		Schemes:  []string{"http"},
	})
}

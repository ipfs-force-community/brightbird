package cmds

import (
	"github.com/ipfs-force-community/brightbird/web/backend/client"
	"github.com/urfave/cli/v2"
)

func DefaulAPI(cliCtx *cli.Context) *client.BrightBirdAPI {
	return client.NewHTTPClientWithConfig(nil, &client.TransportConfig{
		Host:     cliCtx.String("listen"),
		BasePath: "/api/v1",
		Schemes:  []string{"http"},
	})
}

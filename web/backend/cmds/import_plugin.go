package cmds

import (
	"github.com/hunjixin/brightbird/web/backend/client"
	"github.com/hunjixin/brightbird/web/backend/client/operations"
	"github.com/urfave/cli/v2"
)

var ImportPluginsCmds = &cli.Command{
	Name:  "import-plugin",
	Usage: "import plugin mainfest to database",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "path",
			Usage:    "path to read plugins",
			Required: true,
		},
	},
	Action: func(cliCtx *cli.Context) error {
		api := DefaulAPI()
		params := operations.NewImportPluginParamsWithContext(cliCtx.Context)
		params.SetPath(cliCtx.String("path"))
		_, err := api.Operations.ImportPlugin(params)
		return err
	},
}

func DefaulAPI() *client.BrightBirdAPI {
	return client.NewHTTPClientWithConfig(nil, &client.TransportConfig{
		Host:     "127.0.0.1:12356",
		BasePath: "/api/v1",
		Schemes:  []string{"http"},
	})
}

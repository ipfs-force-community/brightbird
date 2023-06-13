package cmds

import (
	"os"

	"github.com/go-openapi/runtime"
	"github.com/hunjixin/brightbird/web/backend/client/plugin"
	"github.com/urfave/cli/v2"
)

var ImportPluginsCmds = &cli.Command{
	Name:      "import-plugin",
	Usage:     "import plugin to database",
	ArgsUsage: "<dir to plugins/file to directory>",
	Flags:     []cli.Flag{},
	Action: func(cliCtx *cli.Context) error {
		api := DefaulAPI()
		path := cliCtx.Args().Get(0)
		st, err := os.Stat(path)
		if err != nil {
			return err
		}

		if st.IsDir() {
			params := plugin.NewImportPluginParamsWithContext(cliCtx.Context)
			params.SetPath(path)
			_, err := api.Plugin.ImportPlugin(params)
			if err != nil {
				return err
			}
		} else {
			params := plugin.NewUploadPluginFilesParamsParamsWithContext(cliCtx.Context)
			reader, err := os.Open(path)
			if err != nil {
				return err
			}

			params.SetPluginFiles(runtime.NamedReader("plugins", reader))
			_, err = api.Plugin.UploadPluginFilesParams(params)
			if err != nil {
				return err
			}
		}

		return nil
	},
}

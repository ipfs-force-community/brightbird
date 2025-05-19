//go:generate swagger generate spec -m -o ./swagger.json
package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/ipfs-force-community/brightbird/pkg/version"
)

func main() {
	app := &cli.App{
		Name:    "backend",
		Usage:   "test plateform backend",
		Version: version.Version(),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "listen",
				EnvVars: []string{"BRIGHTBIRT_BACKEND_LISTEN"},
				Usage:   "listen api address",
				Value:   "127.0.0.1:12356",
			},
		},
		Commands: []*cli.Command{
			ExampleCmd,
			ImportPluginsCmds,
			RunCmd,
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
		return
	}
}

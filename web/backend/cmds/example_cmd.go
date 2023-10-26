package cmds

import (
	"bytes"
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/ipfs-force-community/brightbird/web/backend/config"
	"github.com/urfave/cli/v2"
)

var ExampleCmd = &cli.Command{
	Name: "config",
	Action: func(cliCtx *cli.Context) error {
		cfg := config.Config{}
		buf := new(bytes.Buffer)
		if err := toml.NewEncoder(buf).Encode(cfg); err != nil {
			return err
		}
		tomlData := buf.String()
		fmt.Println(string(tomlData)) //nolint
		return nil
	},
}

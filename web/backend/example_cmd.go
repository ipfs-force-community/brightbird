package main

import (
	"fmt"

	"github.com/pelletier/go-toml/v2"
	"github.com/urfave/cli/v2"
)

var exampleCmd = &cli.Command{
	Name: "config",
	Action: func(_ *cli.Context) error {
		cfg := Config{}
		data, err := toml.Marshal(cfg)
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	},
}

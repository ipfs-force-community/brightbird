package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/filecoin-project/venus-miner/build"
	"github.com/google/uuid"
	"github.com/hunjixin/brightbird/env"
	fx_opt "github.com/hunjixin/brightbird/fx_opt"
	"github.com/hunjixin/brightbird/types"
	logging "github.com/ipfs/go-log/v2"
	"github.com/urfave/cli/v2"
	"go.uber.org/fx"
)

var mainLog = logging.Logger("main")

func main() {
	app := &cli.App{
		Name:    "lotus-health",
		Usage:   "Tools for monitoring lotus daemon health",
		Version: build.UserVersion(),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "config",
				Value:    "",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "testfile",
				Value: "",
			},
			&cli.StringFlag{
				Name:  "plugins",
				Value: "",
			},
			&cli.IntFlag{
				Name:  "timeout",
				Value: 0,
				Usage: "timeout for testing unit(m)",
			},
			&cli.StringFlag{
				Name:  "log-level",
				Value: "INFO",
			},
		},
		Action: func(c *cli.Context) error {
			err := logging.SetLogLevel("*", c.String("log-level"))
			if err != nil {
				return err
			}

			configPath := c.String("config")
			configFileContent, err := os.ReadFile(configPath)
			if err != nil {
				return err
			}
			cfg := Config{}
			err = toml.Unmarshal(configFileContent, &cfg)
			if err != nil {
				return err
			}
			if c.IsSet("testfile") {
				cfg.TestFile = c.String("testfile")
			}
			if c.IsSet("plugins") {
				cfg.PluginStore = c.String("plugins")
			}
			if c.IsSet("timeout") {
				cfg.Timeout = c.Int("timeout")
			}
			return run(c.Context, cfg)
		},
	}

	if err := app.Run(os.Args); err != nil {
		mainLog.Error(err)
		os.Exit(1)
		return
	}
}

func run(ctx context.Context, cfg Config) (err error) {
	content, err := os.ReadFile(cfg.TestFile)
	if err != nil {
		return
	}
	flow := &types.TestFlow{}
	err = json.Unmarshal(content, flow)
	if err != nil {
		return
	}

	execStore, err := types.LoadPlugins(filepath.Join(cfg.PluginStore, "exec"))
	if err != nil {
		return
	}

	deployPlugin, err := types.LoadPlugins(filepath.Join(cfg.PluginStore, "deploy"))
	if err != nil {
		return
	}

	cleaner := Cleaner{}
	defer func() {
		if err := cleaner.DoClean(); err != nil {
			mainLog.Errorf("clean up failed %v", err)
		}
	}()

	stop, err := fx_opt.New(ctx,
		fx_opt.Override(new(context.Context), func(lc fx.Lifecycle) context.Context {
			if cfg.Timeout > 0 {
				tCtx, _ := context.WithTimeout(ctx, time.Minute*time.Duration(cfg.Timeout))
				return tCtx
			}
			return ctx
		}),
		fx_opt.Override(new(types.BootstrapPeers), types.BootstrapPeers(cfg.BootstrapPeers)),
		fx_opt.Override(new(types.TestId), types.TestId(uuid.New().String()[:8])),
		fx_opt.Override(new(*env.K8sEnvDeployer), func(lc fx.Lifecycle, testId types.TestId) (*env.K8sEnvDeployer, error) {
			k8sEnv, err := env.NewK8sEnvDeployer("default", string(testId))
			if err != nil {
				return nil, err
			}

			cleaner.AddFunc(func() error {
				mainLog.Infof("start to cleanup k8s resource")
				return k8sEnv.Clean(ctx)
			})
			return k8sEnv, nil
		}),
		DeployFLow(flow.Nodes, deployPlugin),
		ExecFlow(execStore, flow.Cases),
	)
	if err != nil {
		return
	}
	return stop(ctx)
}

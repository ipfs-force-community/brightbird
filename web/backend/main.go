//go:generate swagger generate spec -m -o ./swagger.json
package main

import (
	"context"
	"github.com/BurntSushi/toml"
	"github.com/gin-gonic/gin"
	"github.com/hunjixin/brightbird/fx_opt"
	"github.com/hunjixin/brightbird/repo"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/utils"
	"github.com/hunjixin/brightbird/version"
	"github.com/hunjixin/brightbird/web/backend/api"
	logging "github.com/ipfs/go-log/v2"
	"github.com/urfave/cli/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net"
	"net/http"
	"os"
	"path/filepath"
)

var log = logging.Logger("main")

func main() {
	app := &cli.App{
		Name:    "lotus-health",
		Usage:   "Tools for monitoring lotus daemon health",
		Version: version.Version(),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "config",
				Value: "",
			},
			&cli.StringFlag{
				Name:  "mongo",
				Value: "mongodb://localhost:27017",
			},
			&cli.StringFlag{
				Name:  "plugins",
				Value: "",
			},
			&cli.StringFlag{
				Name:  "listen",
				Value: "127.0.0.1:12356",
			},
		},
		Action: func(c *cli.Context) error {
			err := logging.SetLogLevel("*", c.String("log-level"))
			if err != nil {
				return err
			}

			cfg := DefaultConfig()
			if c.IsSet("config") {
				configPath := c.String("config")
				configFileContent, err := os.ReadFile(configPath)
				if err != nil {
					return err
				}
				err = toml.Unmarshal(configFileContent, &cfg)
				if err != nil {
					return err
				}
			}

			if c.IsSet("plugins") {
				cfg.PluginStore = c.String("plugins")
			}

			if c.IsSet("listen") {
				cfg.Listen = c.String("listen")
			}

			if c.IsSet("mongo") {
				cfg.MongoUrl = c.String("mongo")
			}
			return run(c.Context, cfg)
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Error(err)
		os.Exit(1)
		return
	}
}

func run(ctx context.Context, cfg Config) error {
	e := gin.Default()
	e.Use(corsMiddleWare())
	e.Use(errorHandleMiddleWare())

	shutdown := make(types.Shutdown)
	stop, err := fx_opt.New(ctx,
		fx_opt.Override(new(*gin.Engine), e),
		fx_opt.Override(new(*api.V1RouterGroup), func(e *gin.Engine) *api.V1RouterGroup {
			return (*api.V1RouterGroup)(e.Group("api/v1"))
		}),
		fx_opt.Override(new(context.Context), ctx),
		fx_opt.Override(new(*mongo.Database), func() (*mongo.Database, error) {
			client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoUrl))
			if err != nil {
				return nil, err
			}
			return client.Database("test-platform"), nil
		}),
		fx_opt.Override(new(repo.DeployPluginStore), func() (repo.DeployPluginStore, error) {
			return types.LoadPlugins(filepath.Join(cfg.PluginStore, "deploy"))
		}),
		fx_opt.Override(new(repo.ExecPluginStore), func() (repo.ExecPluginStore, error) {
			return types.LoadPlugins(filepath.Join(cfg.PluginStore, "exec"))
		}),
		fx_opt.Override(new(repo.IPluginService), NewPlugin),

		//group repo
		fx_opt.Override(new(repo.ITestFlowRepo), NewTestFlowRepo),
		fx_opt.Override(new(repo.IGroupRepo), NewGroupRepo),
		fx_opt.Override(new(repo.IJobRepo), NewJobRepo),
		fx_opt.Override(new(repo.ITaskRepo), NewTaskRepo),

		fx_opt.Override(fx_opt.NextInvoke(), api.RegisterCommonRouter),
		fx_opt.Override(fx_opt.NextInvoke(), api.RegisterDeployRouter),
		fx_opt.Override(fx_opt.NextInvoke(), api.RegisterTestFlowRouter),
		fx_opt.Override(fx_opt.NextInvoke(), api.RegisterGroupRouter),
	)
	if err != nil {
		return err
	}
	go utils.CatchSig(ctx, shutdown)

	listener, err := net.Listen("tcp", cfg.Listen)
	if err != nil {
		return err
	}
	defer listener.Close() //nolint

	log.Infof("Start listen api %s", listener.Addr())
	go func() {
		err = e.RunListener(listener)
		if err != nil {
			log.Errorf("listen address fail %s", err)
		}
	}()
	<-shutdown
	return stop(ctx)
}

func errorHandleMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if c.Errors != nil {
			c.Writer.WriteHeader(http.StatusServiceUnavailable)
			c.Writer.Write([]byte(c.Errors.String()))
		}
	}
}

func corsMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "*")
		c.Header("Access-Control-Allow-Headers", "*")
		c.Header("Content-Type", "*")
		if c.Request.Method == "OPTIONS" {
			c.JSON(http.StatusOK, "ok!")
			return
		}
		c.Next()
	}
}

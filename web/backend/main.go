//go:generate swagger generate spec -m -o ./swagger.json
package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/gin-gonic/gin"
	"github.com/hunjixin/brightbird/fx_opt"
	"github.com/hunjixin/brightbird/repo"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/utils"
	"github.com/hunjixin/brightbird/version"
	"github.com/hunjixin/brightbird/web/backend/api"
	"github.com/hunjixin/brightbird/web/backend/config"
	"github.com/hunjixin/brightbird/web/backend/job"
	logging "github.com/ipfs/go-log/v2"
	"github.com/robfig/cron/v3"
	"github.com/urfave/cli/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/fx"
)

var log = logging.Logger("main")

func main() {
	app := &cli.App{
		Name:    "backend",
		Usage:   "test plateform backend",
		Version: version.Version(),
		Commands: []*cli.Command{
			exampleCmd,
			runCmd,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Error(err)
		os.Exit(1)
		return
	}
}

var runCmd = &cli.Command{
	Name:  "run",
	Usage: "Tools for monitoring lotus daemon health",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "config",
			Value: "",
		},
		&cli.StringFlag{
			Name:  "plugins",
			Value: "",
		},
		&cli.StringFlag{
			Name:  "proxy",
			Value: "",
		},
		&cli.StringFlag{
			Name:  "build-space",
			Value: "",
		},
		&cli.StringFlag{
			Name:  "runner-cfg",
			Value: "",
		},
		&cli.StringFlag{
			Name:  "mongo",
			Value: "mongodb://localhost:27017",
		},
		&cli.StringFlag{
			Name:  "dbName",
			Value: "testplateform",
		},
		&cli.StringFlag{
			Name:  "listen",
			Value: "127.0.0.1:12356",
		},
		&cli.StringFlag{
			Name:  "log-level",
			Value: "debug",
		},
	},
	Action: func(c *cli.Context) error {
		err := logging.SetLogLevel("*", c.String("log-level"))
		if err != nil {
			return err
		}

		cfg := config.DefaultConfig()
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

		if c.IsSet("dbName") {
			cfg.DbName = c.String("dbBane")
		}

		if c.IsSet("mongoUrl") {
			cfg.MongoUrl = c.String("mongoUrl")
		}

		if c.IsSet("proxy") {
			cfg.Proxy = c.String("proxy")
		}

		if c.IsSet("build-space") {
			cfg.BuildSpace = c.String("build-space")
		}

		if c.IsSet("runner-cfg") {
			cfg.RunnerConfig = c.String("runner-cfg")
		}

		return run(c.Context, cfg)
	},
}

func run(pCtx context.Context, cfg config.Config) error {
	e := gin.Default()
	e.Use(corsMiddleWare())
	e.Use(errorHandleMiddleWare())

	shutdown := make(types.Shutdown)
	stop, err := fx_opt.New(pCtx,
		fx_opt.Override(new(context.Context), pCtx),
		//config
		fx_opt.Override(new(config.Config), cfg),
		fx_opt.Override(new(types.PrivateRegistry), NewPrivateRegistry(cfg)),

		fx_opt.Override(new(*gin.Engine), e),
		fx_opt.Override(new(*api.V1RouterGroup), func(e *gin.Engine) *api.V1RouterGroup {
			return (*api.V1RouterGroup)(e.Group("api/v1"))
		}),
		fx_opt.Override(new(*mongo.Database), func(ctx context.Context) (*mongo.Database, error) {
			client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoUrl))
			if err != nil {
				return nil, err
			}
			return client.Database(cfg.DbName), nil
		}),
		fx_opt.Override(new(repo.DeployPluginStore), func() (repo.DeployPluginStore, error) {
			return types.LoadPlugins(filepath.Join(cfg.PluginStore, "deploy"))
		}),
		fx_opt.Override(new(repo.ExecPluginStore), func() (repo.ExecPluginStore, error) {
			return types.LoadPlugins(filepath.Join(cfg.PluginStore, "exec"))
		}),
		//k8s env
		fx_opt.Override(new(*job.TestRunnerDeployer), func(lc fx.Lifecycle) (*job.TestRunnerDeployer, error) {
			return job.NewTestRunnerDeployer("default")
		}),
		//deploy plugin
		fx_opt.Override(new(repo.IPluginService), NewPlugin),

		// build
		fx_opt.Override(new(job.FFIDownloader), NewFFIDownloader(cfg)),
		fx_opt.Override(new(job.IDockerOperation), NewDockerRegistry(cfg)),
		fx_opt.Override(new(job.IBuilderWorkerProvider), NewBuilderWorkerProvidor(cfg)),
		fx_opt.Override(new(*job.ImageBuilderMgr), NewBuilderMgr(cfg)),
		//job
		fx_opt.Override(new(*cron.Cron), NewCron),
		fx_opt.Override(new(*job.TaskMgr), NewTaskMgr(cfg)),
		fx_opt.Override(new(job.IJobManager), NewJobManager),
		//data repo
		fx_opt.Override(new(repo.ITestFlowRepo), NewTestFlowRepo),
		fx_opt.Override(new(repo.IGroupRepo), NewGroupRepo),
		fx_opt.Override(new(repo.IJobRepo), NewJobRepo),
		fx_opt.Override(new(repo.ITaskRepo), NewTaskRepo),

		//use proxy
		fx_opt.Override(fx_opt.NextInvoke(), UseProxy),
		fx_opt.Override(fx_opt.NextInvoke(), UseGitToken),
		//api
		fx_opt.Override(fx_opt.NextInvoke(), api.RegisterCommonRouter),
		fx_opt.Override(fx_opt.NextInvoke(), api.RegisterDeployRouter),
		fx_opt.Override(fx_opt.NextInvoke(), api.RegisterTestFlowRouter),
		fx_opt.Override(fx_opt.NextInvoke(), api.RegisterGroupRouter),
		fx_opt.Override(fx_opt.NextInvoke(), api.RegisterJobRouter),
		fx_opt.Override(fx_opt.NextInvoke(), api.RegisterTaskRouter),
		fx_opt.Override(fx_opt.NextInvoke(), api.RegisterLogRouter),

		//start
		fx_opt.Override(fx_opt.NextInvoke(), func(ctx context.Context, builder *job.ImageBuilderMgr) {
			go builder.Start(ctx)
		}),
		fx_opt.Override(fx_opt.NextInvoke(), func(ctx context.Context, taskMgr *job.TaskMgr) {
			go taskMgr.Start(ctx)
		}),
		fx_opt.Override(fx_opt.NextInvoke(), func(ctx context.Context, jobMgr job.IJobManager) {
			go jobMgr.Start(ctx)
		}),
	)
	if err != nil {
		return err
	}

	go utils.CatchSig(pCtx, shutdown)

	listener, err := net.Listen("tcp", cfg.Listen)
	if err != nil {
		return err
	}
	defer listener.Close() //nolint

	log.Infof("Start listen api %s", listener.Addr())
	go func() {
		err = e.RunListener(listener)
		if err != nil && errors.Is(err, http.ErrServerClosed) {
			log.Errorf("listen address fail %s", err)
		}
	}()
	<-shutdown

	return stop(pCtx)
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

//go:generate swagger generate spec -m -o ./swagger.json
package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/static"
	"github.com/ipfs-force-community/brightbird/models"

	"github.com/BurntSushi/toml"
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v51/github"
	"github.com/ipfs-force-community/brightbird/fx_opt"
	"github.com/ipfs-force-community/brightbird/repo"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"
	"github.com/ipfs-force-community/brightbird/web/backend/api"
	"github.com/ipfs-force-community/brightbird/web/backend/cmds"
	"github.com/ipfs-force-community/brightbird/web/backend/config"
	"github.com/ipfs-force-community/brightbird/web/backend/job"
	"github.com/ipfs-force-community/brightbird/web/backend/modules"
	logging "github.com/ipfs/go-log/v2"
	"github.com/robfig/cron/v3"
	"github.com/urfave/cli/v2"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var log = logging.Logger("main")

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
			cmds.ExampleCmd,
			cmds.ImportPluginsCmds,
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
			Name:  "static-root",
			Value: "dist",
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

		if c.IsSet("static-root") {
			cfg.StaticRoot = c.String("static-root")
		}

		if c.IsSet("plugins") {
			cfg.PluginStore = c.String("plugins")
		}

		if c.IsSet("log-level") {
			cfg.LogLevel = c.String("log-level")
		}

		if c.IsSet("listen") {
			cfg.Listen = c.String("listen")
		}

		if c.IsSet("dbName") {
			cfg.DBName = c.String("dbBane")
		}

		if c.IsSet("mongoUrl") {
			cfg.MongoURL = c.String("mongoUrl")
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
	err := logging.SetLogLevel("*", cfg.LogLevel)
	if err != nil {
		return err
	}
	logLevel, err := logging.LevelFromString(cfg.LogLevel)
	if err != nil {
		return err
	}

	if logLevel > logging.LevelDebug {
		gin.SetMode(gin.ReleaseMode)
	}

	e := gin.Default()
	if len(cfg.StaticRoot) > 0 {
		e.NoRoute(func(ctx *gin.Context) {
			ctx.Redirect(http.StatusTemporaryRedirect, "/index.html")
		})
		e.Use(static.Serve("/", static.LocalFile(cfg.StaticRoot, false)))
	}
	e.Use(gzip.Gzip(gzip.DefaultCompression))
	e.Use(corsMiddleWare())
	e.Use(errorHandleMiddleWare())

	shutdown := make(types.Shutdown)
	stop, err := fx_opt.New(pCtx,
		fx_opt.Override(new(context.Context), pCtx),
		//config
		fx_opt.Override(new(config.Config), cfg),
		fx_opt.Override(new(types.PrivateRegistry), NewPrivateRegistry(cfg)),
		fx_opt.Override(new(types.PluginStore), types.PluginStore(cfg.PluginStore)),
		//webhook
		fx_opt.Override(new(modules.WebHookPubsub), NewWebhoobPubsub(cfg)),
		//database
		fx_opt.Override(new(*mongo.Database), func(ctx context.Context) (*mongo.Database, error) {
			cmdMonitor := &event.CommandMonitor{
				Started: func(_ context.Context, evt *event.CommandStartedEvent) {
					log.Debugf(evt.Command.String())
				},
			}
			client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURL).SetMonitor(cmdMonitor))
			if err != nil {
				return nil, err
			}
			return client.Database(cfg.DBName), nil
		}),
		//k8s env
		fx_opt.Override(new(*job.TestRunnerDeployer), func() (*job.TestRunnerDeployer, error) {
			return job.NewTestRunnerDeployer(cfg.NameSpace, cfg.Mysql, cfg.LogPath)
		}),

		// build
		fx_opt.Override(new(job.IDockerOperation), NewDockerRegistry(cfg)),
		fx_opt.Override(new(job.IBuilderWorkerProvider), NewBuilderWorkerProvidor(cfg)),
		fx_opt.Override(new(*job.ImageBuilderMgr), NewBuilderMgr(cfg)),
		//job
		fx_opt.Override(new(*cron.Cron), NewCron),
		fx_opt.Override(new(*job.TaskMgr), NewTaskMgr(cfg)),
		fx_opt.Override(new(job.IJobManager), NewJobManager),
		//data repo
		fx_opt.Override(new(repo.IPluginService), NewPlugin),
		fx_opt.Override(new(repo.ITestFlowRepo), NewTestFlowRepo),
		fx_opt.Override(new(repo.IGroupRepo), NewGroupRepo),
		fx_opt.Override(new(repo.IJobRepo), NewJobRepo),
		fx_opt.Override(new(repo.ITaskRepo), NewTaskRepo),
		fx_opt.Override(new(repo.ILogRepo), NewLogRepo),

		//use proxy
		fx_opt.Override(fx_opt.NextInvoke(), UseProxy),
		fx_opt.Override(fx_opt.NextInvoke(), UseGitToken),
		fx_opt.Override(new(*github.Client), NewGithubClient(cfg)),
		//api
		fx_opt.Override(new(*gin.Engine), e),
		fx_opt.Override(new(*api.V1RouterGroup), func(e *gin.Engine) *api.V1RouterGroup {
			return e.Group("api/v1")
		}),
		fx_opt.Override(fx_opt.NextInvoke(), api.RegisterCommonRouter),
		fx_opt.Override(fx_opt.NextInvoke(), api.RegisterDeployRouter),
		fx_opt.Override(fx_opt.NextInvoke(), api.RegisterTestFlowRouter),
		fx_opt.Override(fx_opt.NextInvoke(), api.RegisterGroupRouter),
		fx_opt.Override(fx_opt.NextInvoke(), api.RegisterJobRouter),
		fx_opt.Override(fx_opt.NextInvoke(), api.RegisterTaskRouter),
		fx_opt.Override(fx_opt.NextInvoke(), api.RegisterLogRouter),
		fx_opt.Override(fx_opt.NextInvoke(), api.RegisterDashboardRouter),

		//start
		fx_opt.Override(fx_opt.NextInvoke(), func(ctx context.Context, builder *job.ImageBuilderMgr) {
			go builder.Start(ctx) //nolint
		}),
		fx_opt.Override(fx_opt.NextInvoke(), func(ctx context.Context, taskMgr *job.TaskMgr) {
			go taskMgr.Start(ctx) //nolint
		}),
		fx_opt.Override(fx_opt.NextInvoke(), func(ctx context.Context, jobMgr job.IJobManager) {
			go jobMgr.Start(ctx) //nolint
		}),
	)
	if err != nil {
		return err
	}

	go types.CatchSig(pCtx, shutdown)

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
	fmt.Println("graceful shutdown")
	return stop(pCtx)
}

func errorHandleMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if c.Errors != nil && len(c.Errors) > 0 {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, models.APIError{Message: c.Errors.String()})
		}
	}
}

func corsMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "*")
		c.Header("Access-Control-Allow-Headers", "*")
		if c.Request.Method == "OPTIONS" {
			c.JSON(http.StatusOK, "ok!")
			return
		}
		c.Next()
	}
}

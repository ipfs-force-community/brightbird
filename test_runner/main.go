package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/hunjixin/brightbird/env/plugin"

	"github.com/hunjixin/brightbird/models"

	"github.com/gin-gonic/gin"
	"github.com/hunjixin/brightbird/repo"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/BurntSushi/toml"

	"github.com/hunjixin/brightbird/env"
	fx_opt "github.com/hunjixin/brightbird/fx_opt"
	"github.com/hunjixin/brightbird/test_runner/runnerctl"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
	logging "github.com/ipfs/go-log/v2"
	"github.com/urfave/cli/v2"
	"go.uber.org/fx"
)

var log = logging.Logger("main")

func main() {
	app := &cli.App{
		Name:    "test runner",
		Usage:   "Tools for running tests",
		Version: version.Version(),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "config",
				Value: "",
			},
			&cli.StringSliceFlag{
				Name: "bootPeer",
			},
			&cli.StringFlag{
				Name:  "plugins",
				Value: "",
			},
			&cli.StringFlag{
				Name:  "namespace",
				Value: "default",
			},
			&cli.StringFlag{
				Name:  "mysql",
				Usage: "config mysql template xxx%sxxxx",
				Value: "",
			},
			&cli.StringFlag{
				Name:  "mongoUrl",
				Value: "mongodb://localhost:27017",
			},
			&cli.StringFlag{
				Name:  "dbName",
				Value: "testplateform",
			},
			&cli.StringFlag{
				Name:  "tmpPath",
				Value: "",
			},
			&cli.IntFlag{
				Name:  "timeout",
				Value: 0,
				Usage: "timeout for testing unit(m)",
			},
			&cli.StringFlag{
				Name:  "taskId",
				Usage: "test  to running",
			},
			&cli.StringFlag{
				Name:  "privReg",
				Usage: "use private registry",
			},
			&cli.StringFlag{
				Name:  "logLevel",
				Value: "INFO",
			},
			&cli.StringFlag{
				Name:  "listen",
				Value: "0.0.0.0:5682",
			},
		},
		Action: func(c *cli.Context) error {
			err := logging.SetLogLevel("*", c.String("logLevel"))
			if err != nil {
				return err
			}

			cfg := Config{}
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

			cfg.Listen = c.String("listen")

			if c.IsSet("plugins") {
				cfg.PluginStore = c.String("plugins")
			}

			if c.IsSet("timeout") {
				cfg.Timeout = c.Int("timeout")
			}

			if c.IsSet("namespace") {
				cfg.NameSpace = c.String("namespace")
			}

			if c.IsSet("taskId") {
				cfg.TaskId = c.String("taskId")
			}

			if c.IsSet("dbName") {
				cfg.DBName = c.String("dbName")
			}

			if c.IsSet("tmpPath") {
				cfg.TmpPath = c.String("tmpPath")
			}

			if c.IsSet("mongoUrl") {
				cfg.MongoURL = c.String("mongoUrl")
			}

			if c.IsSet("mysql") {
				cfg.Mysql = c.String("mysql")
			}

			if c.IsSet("privReg") {
				cfg.PrivateRegistry = c.String("privReg")
			}

			if c.IsSet("bootPeer") {
				cfg.BootstrapPeers = c.StringSlice("bootPeer")
			}

			if len(cfg.TaskId) == 0 {
				return errors.New("task id must be specific")
			}

			if len(cfg.Mysql) == 0 {
				return errors.New("mysql must set")
			}

			return run(c.Context, &cfg)
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Errorf("run test runner fail %v", err)
		os.Exit(1)
		return
	}
	log.Info("testrunner completed")
	os.Exit(0)
}

func run(pCtx context.Context, cfg *Config) (err error) {
	db, err := getDatabase(pCtx, cfg)
	if err != nil {
		return err
	}
	taskId, err := primitive.ObjectIDFromHex(cfg.TaskId)
	if err != nil {
		return err
	}

	taskRepo, err := repo.NewTaskRepo(pCtx, db)
	if err != nil {
		return err
	}

	pluginRepo, err := repo.NewPluginSvc(pCtx, db)
	if err != nil {
		return err
	}

	testflow, err := getTestFLow(pCtx, db, cfg.TaskId)
	if err != nil {
		return err
	}

	task, err := taskRepo.Get(pCtx, taskId)
	if err != nil {
		return err
	}

	if task.State == models.TempError {
		_ = taskRepo.MarkState(pCtx, taskId, models.Running, "restart")
	}

	cleaner := Cleaner{}
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic when run testrunner %v", r)
		}
		if err != nil {
			_ = taskRepo.MarkState(pCtx, taskId, models.TempError, err.Error())
		} else {
			_ = taskRepo.MarkState(pCtx, taskId, models.Successful, "run successfully")
		}

		//todo get logs
		if cleanErr := cleaner.DoClean(); cleanErr != nil {
			log.Errorf("clean up failed %v", cleanErr)
		}
	}()

	initParams := &env.K8sInitParams{
		Namespace:         cfg.NameSpace,
		TestID:            string(task.TestId),
		PrivateRegistry:   cfg.PrivateRegistry,
		MysqlConnTemplate: cfg.Mysql,
		TmpPath:           cfg.TmpPath,
	}
	initNodes := InitedNode{}
	ctx, shutdown := CatchShutdown(pCtx, cfg.Timeout)
	stop, err := fx_opt.New(ctx,
		fx_opt.Override(new(types.Shutdown), shutdown),
		fx_opt.Override(new(context.Context), ctx),

		// plugin
		fx_opt.Override(new(*models.Task), task),

		//config
		fx_opt.Override(new(*Config), cfg),
		fx_opt.Override(new(types.Endpoint), types.Endpoint(cfg.Listen)),
		fx_opt.Override(new(types.BootstrapPeers), types.BootstrapPeers(cfg.BootstrapPeers)),
		fx_opt.Override(new(types.TestId), func(task *models.Task) types.TestId {
			return task.TestId
		}),

		//database
		fx_opt.Override(new(*mongo.Database), db),
		fx_opt.Override(new(repo.ITaskRepo), taskRepo),
		fx_opt.Override(new(repo.ITestFlowRepo), repo.NewTestFlowRepo),
		//k8s
		fx_opt.Override(new(*env.K8sEnvDeployer), func() (*env.K8sEnvDeployer, error) {
			return env.NewK8sEnvDeployer(*initParams)
		}),
		fx_opt.Override(fx_opt.NextInvoke(), func(lc fx.Lifecycle, k8sEnv *env.K8sEnvDeployer) error {
			cleaner.AddFunc(func() error {
				log.Infof("start to cleanup k8s resource")
				return k8sEnv.Clean(pCtx)
			})
			return nil
		}),

		//api
		fx_opt.Override(new(*gin.Engine), gin.Default()),
		fx_opt.Override(new(*runnerctl.APIController), runnerctl.NewAPIController),
		fx_opt.Override(fx_opt.NextInvoke(), runnerctl.SetupAPI),

		//exec testflow
		DeployFLow(pCtx, initNodes, types.BootstrapPeers(cfg.BootstrapPeers), pluginRepo, cfg.PluginStore, string(task.TestId), initParams, testflow.Nodes),
		ExecFlow(pCtx, initNodes, types.BootstrapPeers(cfg.BootstrapPeers), pluginRepo, cfg.PluginStore, string(task.TestId), initParams, testflow.Cases),
	)
	if err != nil {
		return
	}
	return stop(pCtx)
}

func getDatabase(ctx context.Context, cfg *Config) (*mongo.Database, error) {
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
}

func getTestFLow(ctx context.Context, db *mongo.Database, taskIdStr string) (*models.TestFlow, error) {
	taskId, err := primitive.ObjectIDFromHex(taskIdStr)
	if err != nil {
		return nil, err
	}
	taskRep, err := repo.NewTaskRepo(ctx, db)
	if err != nil {
		return nil, err
	}

	testflowRepo, err := repo.NewTestFlowRepo(ctx, db)
	if err != nil {
		return nil, err
	}
	task, err := taskRep.Get(ctx, taskId)
	if err != nil {
		return nil, err
	}

	testFlow, err := testflowRepo.Get(ctx, &repo.GetTestFlowParams{ID: task.TestFlowId})
	if err != nil {
		return nil, err
	}
	//merge version
	for _, node := range testFlow.Nodes {
		version, ok := task.CommitMap[node.Name]
		if !ok {
			return nil, fmt.Errorf("not found version for deploy %s", node.Name)
		}

		codeVersionProp := plugin.FindCodeVersionProperties(node.Properties)
		if codeVersionProp != nil {
			codeVersionProp.Value = version
		} else {
			node.Properties = append(node.Properties, &types.Property{
				Name:    plugin.CodeVersionPropName,
				Type:    "string",
				Value:   version,
				Require: true,
			})
		}
	}
	return testFlow, nil
}

func CatchShutdown(pCtx context.Context, timeout int) (context.Context, types.Shutdown) {
	innerCtx := pCtx
	if timeout > 0 {
		innerCtx, _ = context.WithTimeout(pCtx, time.Minute*time.Duration(timeout)) //nolint
	}
	shutdown := make(types.Shutdown)
	innerCtx, cancel := context.WithCancel(innerCtx)
	go func() {
		fmt.Println("wait shudown")
		<-shutdown
		cancel()
	}()

	go types.CatchSig(pCtx, shutdown)
	return innerCtx, shutdown
}

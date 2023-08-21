package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"github.com/ipfs-force-community/brightbird/models"

	"github.com/gin-gonic/gin"
	"github.com/ipfs-force-community/brightbird/repo"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/BurntSushi/toml"

	"github.com/ipfs-force-community/brightbird/env"
	fx_opt "github.com/ipfs-force-community/brightbird/fx_opt"
	"github.com/ipfs-force-community/brightbird/test_runner/runnerctl"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"
	logging "github.com/ipfs/go-log/v2"
	"github.com/urfave/cli/v2"
	"go.uber.org/fx"
)

const RunnerStart = "RUNNERSTART"
const RunnerEnd = "RUNNEREND"

var log = logging.Logger("main")

func main() {
	time.Sleep(15 * time.Second)
	app := &cli.App{
		Name:    "test runner",
		Usage:   "Tools for running tests",
		Version: version.Version(),
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
				Name:  "registry",
				Usage: "use private registry",
			},
			&cli.StringFlag{
				Name:  "listen",
				Value: "0.0.0.0:5682",
			},
			&cli.StringFlag{
				Name:  "globalParams",
				Value: "{}",
			},
		},
		Action: func(c *cli.Context) error {
			fmt.Println(RunnerStart)
			defer fmt.Println(RunnerEnd) //NOTICE do not remove this code, this is labels to split log

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

			if c.IsSet("mongoUrl") {
				cfg.MongoURL = c.String("mongoUrl")
			}

			if c.IsSet("mysql") {
				cfg.Mysql = c.String("mysql")
			}

			if c.IsSet("registry") {
				cfg.Registry = c.String("registry")
			}

			if c.IsSet("globalParams") {
				var val env.GlobalParams
				err := json.Unmarshal([]byte(c.String("globalParams")), &val)
				if err != nil {
					return err
				}

				cfg.GlobalParams = val
			}

			if len(cfg.TaskId) == 0 {
				return errors.New("task id must be specific")
			}

			if len(cfg.Mysql) == 0 {
				return errors.New("mysql must set")
			}

			err := logging.SetLogLevel("*", c.String("logLevel"))
			if err != nil {
				return err
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

	task, err := taskRepo.Get(pCtx, &repo.GetTaskReq{ID: taskId})
	if err != nil {
		return err
	}

	if task.State == models.TempError {
		_ = taskRepo.MarkState(pCtx, taskId, models.Running, "restart")
	}

	cleaner := Cleaner{}
	defer func() {
		if r := recover(); r != nil {
			reason := fmt.Sprintf("%v", r)
			if val, ok := r.(error); ok {
				reason = val.Error()
			}
			err = fmt.Errorf("panic when run testrunner reason %s stack %s", reason, string(debug.Stack()))
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

	ctx, shutdown := CatchShutdown(pCtx, cfg.Timeout)
	stop, err := fx_opt.New(ctx,
		fx_opt.Override(new(types.Shutdown), shutdown),
		fx_opt.Override(new(context.Context), ctx),

		// plugin
		fx_opt.Override(new(*models.Task), task),
		fx_opt.Override(new(*models.TestFlow), testflow),

		//config
		fx_opt.Override(new(*Config), cfg),
		fx_opt.Override(new(types.Endpoint), types.Endpoint(cfg.Listen)),
		fx_opt.Override(new(types.TestId), func(task *models.Task) types.TestId {
			return task.TestId
		}),

		//database
		fx_opt.Override(new(*mongo.Database), db),
		fx_opt.Override(new(repo.IPluginService), pluginRepo),
		fx_opt.Override(new(repo.ITaskRepo), taskRepo),
		fx_opt.Override(new(repo.ITestFlowRepo), repo.NewTestFlowRepo),
		//k8s
		fx_opt.Override(new(*env.K8sEnvDeployer), func(initParams *env.K8sInitParams) (*env.K8sEnvDeployer, error) {
			return env.NewK8sEnvDeployer(*initParams)
		}),
		fx_opt.Override(new(*env.K8sInitParams), func(task *models.Task, cfg *Config) *env.K8sInitParams {
			return &env.K8sInitParams{
				Namespace:         cfg.NameSpace,
				TestID:            string(task.TestId),
				Registry:          cfg.Registry,
				MysqlConnTemplate: cfg.Mysql,
			}
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
		fx_opt.Override(fx_opt.NextInvoke(), runGraph),
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
	task, err := taskRep.Get(ctx, &repo.GetTaskReq{ID: taskId})
	if err != nil {
		return nil, err
	}

	testFlow, err := testflowRepo.Get(ctx, &repo.GetTestFlowParams{ID: task.TestFlowId})
	if err != nil {
		return nil, err
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

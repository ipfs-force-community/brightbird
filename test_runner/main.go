package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hunjixin/brightbird/repo"
	"github.com/hunjixin/brightbird/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
				Name:     "config",
				Value:    "",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "plugins",
				Value: "",
			},
			&cli.StringFlag{
				Name:  "mysql",
				Usage: "config mysql template xxx%sxxxx",
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
				Name:  "priv_reg",
				Usage: "use private registry",
			},
			&cli.StringFlag{
				Name:  "log-level",
				Value: "INFO",
			},
			&cli.StringFlag{
				Name:  "listen",
				Value: "0.0.0.0:5682",
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

			cfg.Listen = c.String("listen")

			if c.IsSet("plugins") {
				cfg.PluginStore = c.String("plugins")
			}

			if c.IsSet("timeout") {
				cfg.Timeout = c.Int("timeout")
			}

			if c.IsSet("taskId") {
				cfg.TaskId = c.String("taskId")
			}

			if c.IsSet("dbName") {
				cfg.DbName = c.String("dbBane")
			}

			if c.IsSet("mongoUrl") {
				cfg.MongoUrl = c.String("mongoUrl")
			}

			if c.IsSet("mysql") {
				cfg.Mysql = c.String("mysql")
			}

			if c.IsSet("priv_reg") {
				cfg.PrivateRegistry = c.String("priv_reg")
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

	taskRep := repo.NewTaskRepo(db)

	testflow, err := getTestFLow(pCtx, db, cfg.TaskId)
	if err != nil {
		return err
	}

	cleaner := Cleaner{}
	defer func() {
		if err != nil {
			_ = taskRep.MarkState(pCtx, taskId, types.TempError, err.Error())
		} else {
			_ = taskRep.MarkState(pCtx, taskId, types.Successful, "run successfully")
		}

		//todo get logs
		if cleanErr := cleaner.DoClean(); cleanErr != nil {
			log.Errorf("clean up failed %v", cleanErr)
		}
	}()

	deployPlugin, err := types.LoadPlugins(filepath.Join(cfg.PluginStore, "deploy"))
	if err != nil {
		return err
	}

	execPlugin, err := types.LoadPlugins(filepath.Join(cfg.PluginStore, "exec"))
	if err != nil {
		return err
	}
	ctx, shutdown := CatchShutdown(pCtx, cfg.Timeout)
	stop, err := fx_opt.New(ctx,
		fx_opt.Override(new(types.Shutdown), shutdown),
		fx_opt.Override(new(context.Context), ctx),

		// plugin
		fx_opt.Override(new(repo.DeployPluginStore), deployPlugin),
		fx_opt.Override(new(repo.ExecPluginStore), execPlugin),
		fx_opt.Override(new(*types.Task), func(ctx context.Context, taskRepo repo.ITaskRepo) (*types.Task, error) {
			taskId, err := primitive.ObjectIDFromHex(cfg.TaskId)
			if err != nil {
				return nil, err
			}
			return taskRepo.Get(ctx, taskId)
		}),

		//config
		fx_opt.Override(new(*Config), cfg),
		fx_opt.Override(new(types.Endpoint), types.Endpoint(cfg.Listen)),
		fx_opt.Override(new(types.BootstrapPeers), types.BootstrapPeers(cfg.BootstrapPeers)),
		fx_opt.Override(new(types.TestId), func(task *types.Task) types.TestId {
			if len(task.TestId) > 0 {
				return task.TestId
			}
			return types.TestId(uuid.New().String()[:8])
		}),

		//database
		fx_opt.Override(new(*mongo.Database), db),
		fx_opt.Override(new(repo.ITaskRepo), taskRep),
		fx_opt.Override(new(repo.ITestFlowRepo), repo.NewTestFlowRepo),
		//k8s
		fx_opt.Override(new(*env.K8sEnvDeployer), func(lc fx.Lifecycle, testId types.TestId) (*env.K8sEnvDeployer, error) {
			k8sEnv, err := env.NewK8sEnvDeployer("default", string(testId), cfg.PrivateRegistry, cfg.Mysql)
			if err != nil {
				return nil, err
			}

			cleaner.AddFunc(func() error {
				log.Infof("start to cleanup k8s resource")
				return k8sEnv.Clean(pCtx)
			})
			return k8sEnv, nil
		}),

		//api
		fx_opt.Override(new(*gin.Engine), gin.Default()),
		fx_opt.Override(new(*runnerctl.APIController), runnerctl.NewAPIController),
		fx_opt.Override(fx_opt.NextInvoke(), runnerctl.SetupAPI),

		//exec testflow
		DeployFLow(deployPlugin, testflow.Nodes),
		ExecFlow(execPlugin, testflow.Cases),
	)
	if err != nil {
		return
	}
	return stop(pCtx)
}

func getDatabase(ctx context.Context, cfg *Config) (*mongo.Database, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoUrl))
	if err != nil {
		return nil, err
	}
	return client.Database(cfg.DbName), nil
}

func getTestFLow(ctx context.Context, db *mongo.Database, taskIdStr string) (*types.TestFlow, error) {
	taskId, err := primitive.ObjectIDFromHex(taskIdStr)
	if err != nil {
		return nil, err
	}
	taskRep := repo.NewTaskRepo(db)
	testflowRepo := repo.NewTestFlowRepo(db)

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
		version, ok := task.Versions[node.Name]
		if !ok {
			return nil, fmt.Errorf("not found version for deploy %s", node.Name)
		}

		codeVersionProp := findCodeVersionProperties(node.Properties)
		if codeVersionProp != nil {
			codeVersionProp.Value = version
		} else {
			node.Properties = append(node.Properties, &types.Property{
				Name:    types.CodeVersion,
				Type:    "string",
				Value:   version,
				Require: true,
			})
		}
	}
	return testFlow, nil
}

func findCodeVersionProperties(properties []*types.Property) *types.Property {
	for _, property := range properties {
		if property.Name == types.CodeVersion {
			return property
		}
	}
	return nil
}

func CatchShutdown(pCtx context.Context, timeout int) (context.Context, types.Shutdown) {
	innerCtx := pCtx
	if timeout > 0 {
		innerCtx, _ = context.WithTimeout(pCtx, time.Minute*time.Duration(timeout))
	}
	shutdown := make(types.Shutdown)
	innerCtx, cancel := context.WithCancel(innerCtx)
	go func() {
		fmt.Println("wait shudown")
		<-shutdown
		cancel()
	}()

	go utils.CatchSig(pCtx, shutdown)
	return innerCtx, shutdown
}

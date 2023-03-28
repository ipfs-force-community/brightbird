package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/hunjixin/brightbird/repo"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/BurntSushi/toml"
	"github.com/google/uuid"
	"github.com/hunjixin/brightbird/env"
	fx_opt "github.com/hunjixin/brightbird/fx_opt"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
	logging "github.com/ipfs/go-log/v2"
	"github.com/urfave/cli/v2"
	"go.uber.org/fx"
)

var log = logging.Logger("main")

func main() {
	app := &cli.App{
		Name:    "lotus-health",
		Usage:   "Tools for monitoring lotus daemon health",
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
			if c.IsSet("plugins") {
				cfg.PluginStore = c.String("plugins")
			}
			if c.IsSet("timeout") {
				cfg.Timeout = c.Int("timeout")
			}
			if c.IsSet("taskId") {
				cfg.TaskId = c.String("testFlowId")
			}

			if c.IsSet("dbName") {
				cfg.DbName = c.String("dbBane")
			}

			if c.IsSet("mongoUrl") {
				cfg.MongoUrl = c.String("mongoUrl")
			}

			if len(cfg.TaskId) == 0 {
				return errors.New("test flow id must be specific")
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

func run(ctx context.Context, cfg Config) (err error) {
	db, err := getDatabase(ctx, cfg.MongoUrl, cfg.DbName)
	if err != nil {
		return
	}

	flow, err := getTestFLow(ctx, db, cfg.TaskId)
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
		if err != nil {
			_ = markFailTask(ctx, db, cfg.TaskId, err)
		}
		//todo get logs
		if cleanErr := cleaner.DoClean(); cleanErr != nil {
			log.Errorf("clean up failed %v", cleanErr)
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
				log.Infof("start to cleanup k8s resource")
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

func getDatabase(ctx context.Context, mongoUrl string, dbName string) (*mongo.Database, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoUrl))
	if err != nil {
		return nil, err
	}
	return client.Database(dbName), nil
}

func markFailTask(ctx context.Context, db *mongo.Database, taskIdStr string, inErr error) error {
	taskId, err := primitive.ObjectIDFromHex(taskIdStr)
	if err != nil {
		return err
	}

	taskRep := repo.NewTaskRepo(db)
	return taskRep.MarkFail(ctx, taskId, inErr.Error())
}

func getTestFLow(ctx context.Context, db *mongo.Database, taskIdStr string) (*types.TestFlow, error) {
	taskId, err := primitive.ObjectIDFromHex(taskIdStr)
	if err != nil {
		return nil, err
	}
	taskRep := repo.NewTaskRepo(db)
	testflowRepo := repo.NewTestFlowRepo(db, nil)

	task, err := taskRep.Get(ctx, taskId)
	if err != nil {
		return nil, err
	}

	testFlow, err := testflowRepo.GetById(ctx, task.TestFlowId)
	if err != nil {
		return nil, err
	}
	//merge version
	for _, node := range testFlow.Nodes {
		for _, property := range node.Properties {
			if property.Name == types.CodeVersion {
				version, ok := task.Versions[node.Name]
				if !ok {
					return nil, fmt.Errorf("not found version for deploy %s", node.Name)
				}
				property.Value = version
			}
		}
	}
	return testFlow, nil
}

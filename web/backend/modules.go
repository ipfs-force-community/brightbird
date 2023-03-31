package main

import (
	"net/url"
	"os"

	"github.com/hunjixin/brightbird/repo"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/web/backend/config"
	"github.com/hunjixin/brightbird/web/backend/job"
	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/mongo"
)

func UseProxy(cfg config.Config) error {
	err := os.Setenv("ALL_PROXY", cfg.Proxy)
	if err != nil {
		return err
	}
	err = os.Setenv("HTTP_PROXY", cfg.Proxy)
	if err != nil {
		return err
	}
	return os.Setenv("HTTPS_PROXY", cfg.Proxy)
}

func UseGitToken(cfg config.Config) error {
	return os.Setenv("GITHUB_TOKEN", cfg.GitToken)
}

func NewTestFlowRepo(db *mongo.Database) repo.ITestFlowRepo {
	return repo.NewTestFlowRepo(db)
}

func NewGroupRepo(db *mongo.Database, testflowSvc repo.ITestFlowRepo) repo.IGroupRepo {
	return repo.NewGroupSvc(db, testflowSvc)
}

func NewJobRepo(db *mongo.Database) repo.IJobRepo {
	return repo.NewJobRepo(db)
}

func NewTaskRepo(db *mongo.Database) repo.ITaskRepo {
	return repo.NewTaskRepo(db)
}

func NewPlugin(deployPluginStore repo.DeployPluginStore, execPluginStore repo.ExecPluginStore) repo.IPluginService {
	return repo.NewPluginSvc(deployPluginStore, execPluginStore)
}

func NewCron() *cron.Cron {
	return cron.New(cron.WithLogger(job.NewCronLog()))
}

func NewPrivateRegistry(cfg config.Config) func() (types.PrivateRegistry, error) {
	return func() (types.PrivateRegistry, error) {
		regUrl := ""
		if len(cfg.DockerRegistry) == 1 {
			regUrl = cfg.DockerRegistry[0].URL
		} else {
			for _, reg := range cfg.DockerRegistry {
				if reg.Push {
					regUrl = reg.URL
					break
				}
			}
		}
		if len(regUrl) == 0 {
			return "", nil
		}

		url, err := url.Parse(regUrl)
		if err != nil {
			return "", err
		}
		return types.PrivateRegistry(url.Host), nil
	}
}

func NewBuilderMgr(cfg config.Config) func(job.IDockerOperation, repo.DeployPluginStore, types.PrivateRegistry) (*job.ImageBuilderMgr, error) {
	return func(dockerOp job.IDockerOperation, store repo.DeployPluginStore, privateReg types.PrivateRegistry) (*job.ImageBuilderMgr, error) {

		return job.NewImageBuilderMgr(dockerOp, store, cfg.BuildSpace, cfg.Proxy, cfg.GitToken, privateReg), nil
	}
}

func NewJobManager(cron *cron.Cron, taskRepo repo.ITaskRepo, jobRepo repo.IJobRepo) job.IJobManager {
	return job.NewJobManager(cron, taskRepo, jobRepo)
}

func NewTaskMgr(cfg config.Config) func(*cron.Cron, repo.IJobRepo, repo.ITaskRepo, repo.ITestFlowRepo, *job.TestRunnerDeployer, *job.ImageBuilderMgr, types.PrivateRegistry) *job.TaskMgr {
	return func(c *cron.Cron, jobRepo repo.IJobRepo, taskRepo repo.ITaskRepo, testFlowRepo repo.ITestFlowRepo, testRunner *job.TestRunnerDeployer, imageBuilder *job.ImageBuilderMgr, privateReg types.PrivateRegistry) *job.TaskMgr {
		return job.NewTaskMgr(c, jobRepo, taskRepo, testFlowRepo, testRunner, imageBuilder, cfg.RunnerConfig, privateReg)
	}
}

func NewDockerRegistry(cfg config.Config) func() (job.IDockerOperation, error) {
	return func() (job.IDockerOperation, error) {
		return job.NewDockerRegistry(cfg.DockerRegistry)
	}
}

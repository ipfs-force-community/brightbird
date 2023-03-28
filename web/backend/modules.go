package main

import (
	"github.com/hunjixin/brightbird/repo"
	"github.com/hunjixin/brightbird/web/backend/config"
	"github.com/hunjixin/brightbird/web/backend/job"
	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/mongo"
	"net/url"
	"os"
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

func NewTestFlowRepo(db *mongo.Database, execPluginStore repo.ExecPluginStore) repo.ITestFlowRepo {
	return repo.NewTestFlowRepo(db, execPluginStore)
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

func NewPlugin(deployPluginStore repo.DeployPluginStore) repo.IPluginService {
	return repo.NewPluginSvc(deployPluginStore)
}

func NewCron() *cron.Cron {
	return cron.New(cron.WithLogger(job.NewCronLog()))
}

func NewBuilderMgr(cfg config.Config) func(job.IDockerOperation, repo.DeployPluginStore) (*job.ImageBuilderMgr, error) {
	return func(dockerOp job.IDockerOperation, store repo.DeployPluginStore) (*job.ImageBuilderMgr, error) {
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

		url, err := url.Parse(regUrl)
		if err != nil {
			return nil, err
		}
		return job.NewImageBuilderMgr(dockerOp, store, cfg.BuildSpace, cfg.Proxy, cfg.GitToken, url.Host), nil
	}
}

func NewJobManager(cron *cron.Cron, taskRepo repo.ITaskRepo, jobRepo repo.IJobRepo) job.IJobManager {
	return job.NewJobManager(cron, taskRepo, jobRepo)
}

func NewTaskMgr(cfg config.Config) func(*cron.Cron, repo.IJobRepo, repo.ITaskRepo, repo.ITestFlowRepo, *job.TestRunnerDeployer, *job.ImageBuilderMgr) *job.TaskMgr {
	return func(c *cron.Cron, jobRepo repo.IJobRepo, taskRepo repo.ITaskRepo, testFlowRepo repo.ITestFlowRepo, testRunner *job.TestRunnerDeployer, imageBuilder *job.ImageBuilderMgr) *job.TaskMgr {
		return job.NewTaskMgr(c, jobRepo, taskRepo, testFlowRepo, testRunner, imageBuilder, cfg.RunnerConfig)
	}
}

func NewDockerRegistry(cfg config.Config) func() (job.IDockerOperation, error) {
	return func() (job.IDockerOperation, error) {
		return job.NewDockerRegistry(cfg.DockerRegistry)
	}
}

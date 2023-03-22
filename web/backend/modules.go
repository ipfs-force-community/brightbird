package main

import (
	"github.com/hunjixin/brightbird/repo"
	"github.com/hunjixin/brightbird/web/backend/job"
	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/mongo"
)

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

func NewBuilderMgr(cfg Config) func(store repo.DeployPluginStore) *job.ImageBuilderMgr {
	return func(store repo.DeployPluginStore) *job.ImageBuilderMgr {
		return job.NewImageBuilderMgr(store, cfg.BuildSpace, cfg.Proxy)
	}
}

func NewJobManager(cron *cron.Cron, taskRepo repo.ITaskRepo, jobRepo repo.IJobRepo) job.IJobManager {
	return job.NewJobManager(cron, taskRepo, jobRepo)
}

func NewTaskMgr(cfg Config) func(*cron.Cron, repo.IJobRepo, repo.ITaskRepo, repo.ITestFlowRepo, *job.TestRunnerDeployer, *job.ImageBuilderMgr) *job.TaskMgr {
	return func(c *cron.Cron, jobRepo repo.IJobRepo, taskRepo repo.ITaskRepo, testFlowRepo repo.ITestFlowRepo, testRunner *job.TestRunnerDeployer, imageBuilder *job.ImageBuilderMgr) *job.TaskMgr {
		return job.NewTaskMgr(c, jobRepo, taskRepo, testFlowRepo, testRunner, imageBuilder, cfg.RunnerConfig)
	}
}

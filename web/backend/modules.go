package main

import (
	"context"
	"net/url"
	"os"

	"github.com/cskr/pubsub"
	"github.com/google/go-github/v51/github"
	"github.com/hunjixin/brightbird/repo"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/web/backend/config"
	"github.com/hunjixin/brightbird/web/backend/job"
	"github.com/hunjixin/brightbird/web/backend/modules"
	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/oauth2"
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

func NewWebhoobPubsub() modules.WebHookPubsub {
	return pubsub.New(10)
}

func NewTestFlowRepo(ctx context.Context, db *mongo.Database) (repo.ITestFlowRepo, error) {
	return repo.NewTestFlowRepo(ctx, db)
}

func NewGroupRepo(ctx context.Context, db *mongo.Database, testflowSvc repo.ITestFlowRepo) (repo.IGroupRepo, error) {
	return repo.NewGroupSvc(ctx, db, testflowSvc)
}

func NewJobRepo(ctx context.Context, db *mongo.Database) (repo.IJobRepo, error) {
	return repo.NewJobRepo(ctx, db)
}

func NewTaskRepo(ctx context.Context, db *mongo.Database) (repo.ITaskRepo, error) {
	return repo.NewTaskRepo(ctx, db)
}

func NewLogRepo(ctx context.Context, db *mongo.Database) (repo.ILogRepo, error) {
	return repo.NewLogRepo(ctx, db)
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

func NewFFIDownloader(cfg config.Config) func() job.FFIDownloader {
	return func() job.FFIDownloader {
		return job.NewFFIDownloader(cfg.GitToken)
	}
}

func NewBuilderWorkerProvidor(cfg config.Config) func(job.IDockerOperation, job.FFIDownloader, repo.DeployPluginStore, types.PrivateRegistry) (job.IBuilderWorkerProvider, error) {
	return func(dockerOp job.IDockerOperation, ffi job.FFIDownloader, store repo.DeployPluginStore, privateReg types.PrivateRegistry) (job.IBuilderWorkerProvider, error) {
		return job.NewBuildWorkerProvider(dockerOp, store, ffi, privateReg, cfg.Proxy), nil
	}
}

func NewBuilderMgr(cfg config.Config) func(repo.DeployPluginStore, job.IBuilderWorkerProvider) (*job.ImageBuilderMgr, error) {
	return func(store repo.DeployPluginStore, provider job.IBuilderWorkerProvider) (*job.ImageBuilderMgr, error) {
		return job.NewImageBuilderMgr(store, provider, cfg.BuildWorkers), nil
	}
}

func NewJobManager(cron *cron.Cron, deployStore repo.DeployPluginStore, bus modules.WebHookPubsub, githubClient *github.Client, taskRepo repo.ITaskRepo, jobRepo repo.IJobRepo) job.IJobManager {
	return job.NewJobManager(cron, deployStore, bus, githubClient, taskRepo, jobRepo)
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

func NewGithubClient(cfg config.Config) func(context.Context) (*github.Client, error) {
	return func(ctx context.Context) (*github.Client, error) {
		return github.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: cfg.GitToken}))), nil
	}
}

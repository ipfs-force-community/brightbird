package main

import (
	"context"
	"net/url"
	"os"

	"github.com/cskr/pubsub"
	"github.com/google/go-github/v51/github"
	"github.com/hunjixin/brightbird/hookforward/webhooklisten"
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

func NewWebhoobPubsub(cfg config.Config) func(ctx context.Context) (modules.WebHookPubsub, error) {
	return func(ctx context.Context) (modules.WebHookPubsub, error) {
		webhookPubsub := pubsub.New(10)
		ch, err := webhooklisten.WaitForWebHookEvent(ctx, cfg.WebhookUrl)
		if err != nil {
			return nil, err
		}
		go func() {
			for gitEvent := range ch {
				eventType := gitEvent.Header.Get(github.EventTypeHeader)
				event, err := github.ParseWebHook(eventType, gitEvent.Body)
				if err != nil {
					log.Info("parse event fail drop event")
					continue
				}

				if eventType == "release" {
					releaseEvent := event.(*github.ReleaseEvent)
					if releaseEvent.GetRelease().TagName != nil { //commit or tags
						webhookPubsub.Pub(releaseEvent, modules.RELEASE_TOPIC)
						continue
					}
				}

				if eventType == "pull_request" {
					prEvent := event.(*github.PullRequestEvent)
					if prEvent.GetPullRequest().GetMerged() && prEvent.GetPullRequest().GetState() == "closed" {
						webhookPubsub.Pub(event, modules.PR_MERGED_TOPIC)
						continue
					}
				}
			}
		}()
		return webhookPubsub, nil
	}
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

func NewPlugin(ctx context.Context, db *mongo.Database) (repo.IPluginService, error) {
	return repo.NewPluginSvc(ctx, db)
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

func NewBuilderWorkerProvidor(cfg config.Config) func(job.IDockerOperation, job.FFIDownloader, repo.IPluginService, types.PrivateRegistry) (job.IBuilderWorkerProvider, error) {
	return func(dockerOp job.IDockerOperation, ffi job.FFIDownloader, pluginRepo repo.IPluginService, privateReg types.PrivateRegistry) (job.IBuilderWorkerProvider, error) {
		return job.NewBuildWorkerProvider(dockerOp, pluginRepo, ffi, privateReg, cfg.Proxy), nil
	}
}

func NewBuilderMgr(cfg config.Config) func(repo.IPluginService, job.IBuilderWorkerProvider) (*job.ImageBuilderMgr, error) {
	return func(pluginRepo repo.IPluginService, provider job.IBuilderWorkerProvider) (*job.ImageBuilderMgr, error) {
		return job.NewImageBuilderMgr(pluginRepo, provider, cfg.BuildWorkers), nil
	}
}

func NewJobManager(cron *cron.Cron, pluginRepo repo.IPluginService, bus modules.WebHookPubsub, githubClient *github.Client, taskRepo repo.ITaskRepo, jobRepo repo.IJobRepo, testflowRepo repo.ITestFlowRepo) job.IJobManager {
	return job.NewJobManager(cron, pluginRepo, bus, githubClient, taskRepo, jobRepo, testflowRepo)
}

func NewTaskMgr(cfg config.Config) func(*cron.Cron, repo.IJobRepo, repo.ITaskRepo, repo.ITestFlowRepo, *job.TestRunnerDeployer, *job.ImageBuilderMgr, types.PrivateRegistry) *job.TaskMgr {
	return func(c *cron.Cron, jobRepo repo.IJobRepo, taskRepo repo.ITaskRepo, testFlowRepo repo.ITestFlowRepo, testRunner *job.TestRunnerDeployer, imageBuilder *job.ImageBuilderMgr, privateReg types.PrivateRegistry) *job.TaskMgr {
		return job.NewTaskMgr(cfg, c, jobRepo, taskRepo, testFlowRepo, testRunner, imageBuilder, cfg.RunnerConfig, privateReg)
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

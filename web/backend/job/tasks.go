package job

import (
	"context"
	"os"
	"time"

	"github.com/hunjixin/brightbird/repo"
	"github.com/hunjixin/brightbird/types"
	"github.com/robfig/cron/v3"
)

type TaskMgr struct {
	c            *cron.Cron
	jobRepo      repo.IJobRepo
	taskRepo     repo.ITaskRepo
	testFlowRepo repo.ITestFlowRepo
	testRunner   *TestRunnerDeployer
	imageBuilder *ImageBuilderMgr
	runnerConfig string
}

func NewTaskMgr(c *cron.Cron, jobRepo repo.IJobRepo, taskRepo repo.ITaskRepo, testFlowRepo repo.ITestFlowRepo, testRunner *TestRunnerDeployer, imageBuilder *ImageBuilderMgr, runnerConfig string) *TaskMgr {
	return &TaskMgr{
		c:            c,
		jobRepo:      jobRepo,
		taskRepo:     taskRepo,
		testFlowRepo: testFlowRepo,
		testRunner:   testRunner,
		imageBuilder: imageBuilder,
		runnerConfig: runnerConfig,
	}
}

func (taskMgr *TaskMgr) Start(ctx context.Context) error {
	tm := time.NewTicker(time.Minute)
	defer tm.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-tm.C:
			//scan tasks to process
			jobs, err := taskMgr.jobRepo.List(ctx) //todo 数据规模大了 可以考虑采用被动触发的方式 现在这种做法简单一些
			if err != nil {
				log.Error("fetch job list fail %v", err)
				continue
			}
			for _, job := range jobs {
				tasks, err := taskMgr.taskRepo.List(ctx, repo.ListParams{JobId: job.ID})
				if err != nil {
					log.Error("fetch task list fail %v", err)
					continue
				}
				for _, task := range tasks {
					err = taskMgr.Process(ctx, task)
					if err != nil {
						log.Errorf("process task (%s) fail %v", task.ID, err)
						continue
					}
				}
			}
		}
	}
}

func (taskMgr *TaskMgr) Process(ctx context.Context, task *types.Task) error {
	job, err := taskMgr.jobRepo.Get(ctx, task.JobId)
	if err != nil {
		return err
	}
	testFlow, err := taskMgr.testFlowRepo.GetById(ctx, job.TestFlowId)
	if err != nil {
		log.Errorf("get test flow failed %v", err)
		return err
	}

	//confirm version and build image.
	versionMap, err := taskMgr.imageBuilder.BuildTestFlowEnv(ctx, testFlow.Nodes, job.Versions) //todo maybe move this code to previous step
	if err != nil {
		return err
	}

	//save testflow as task params
	err = taskMgr.taskRepo.UpdateVersion(ctx, task.ID, versionMap)
	if err != nil {
		return err
	}

	//run test flow
	file, err := os.Open(taskMgr.runnerConfig)
	if err != nil {
		return err
	}

	return taskMgr.testRunner.ApplyRunner(ctx, file, map[string]string{
		"TestFlowId": job.TestFlowId.String(),
		"TestId":     string(task.TestId),
	})
}

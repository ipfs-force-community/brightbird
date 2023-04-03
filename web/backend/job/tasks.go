package job

import (
	"context"
	"fmt"
	"os"
	"time"

	errors2 "k8s.io/apimachinery/pkg/api/errors"

	"github.com/hunjixin/brightbird/repo"
	"github.com/hunjixin/brightbird/types"
	logging "github.com/ipfs/go-log/v2"
	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/bson/primitive"
	corev1 "k8s.io/api/core/v1"
)

var taskLog = logging.Logger("task")

type TaskMgr struct {
	c            *cron.Cron
	jobRepo      repo.IJobRepo
	taskRepo     repo.ITaskRepo
	testFlowRepo repo.ITestFlowRepo
	testRunner   *TestRunnerDeployer
	imageBuilder *ImageBuilderMgr

	privateRegistry types.PrivateRegistry
	runnerConfig    string
}

func NewTaskMgr(c *cron.Cron, jobRepo repo.IJobRepo, taskRepo repo.ITaskRepo, testFlowRepo repo.ITestFlowRepo, testRunner *TestRunnerDeployer, imageBuilder *ImageBuilderMgr, runnerConfig string, privateReg types.PrivateRegistry) *TaskMgr {
	return &TaskMgr{
		c:               c,
		jobRepo:         jobRepo,
		taskRepo:        taskRepo,
		testFlowRepo:    testFlowRepo,
		testRunner:      testRunner,
		imageBuilder:    imageBuilder,
		runnerConfig:    runnerConfig,
		privateRegistry: privateReg,
	}
}

func (taskMgr *TaskMgr) Start(ctx context.Context) error {
	tm := time.NewTicker(time.Minute)
	defer tm.Stop()

	for {
		//scan tasks to process
		jobs, err := taskMgr.jobRepo.List(ctx) //todo 数据规模大了 可以考虑采用被动触发的方式 现在这种做法简单一些
		if err != nil {
			taskLog.Error("fetch job list fail %v", err)
			continue
		}
		for _, job := range jobs {
			//check running task state
			runningTask, err := taskMgr.taskRepo.List(ctx, repo.ListParams{JobId: job.ID, State: []types.State{types.Running}})
			if err != nil {
				taskLog.Errorf("fetch running task list fail %v", err)
				continue
			}
			for _, task := range runningTask {
				restartCount, err := taskMgr.testRunner.CheckTestRunner(ctx, task.PodName)
				if err != nil {
					if errors2.IsNotFound(err) {
						markFailErr := taskMgr.taskRepo.MarkState(ctx, task.ID, types.Error, err.Error())
						if markFailErr != nil {
							log.Errorf("cannot mark task as fail origin err %v %v", err, markFailErr)
						}
						continue
					} else {
						if restartCount > 5 {
							// mark pod as fail and remove this pod
							markFailErr := taskMgr.taskRepo.MarkState(ctx, task.ID, types.Error, err.Error())
							if err != nil {
								log.Errorf("cannot mark task as fail origin err %v %v", err, markFailErr)
							}

							cleanK8sErr := taskMgr.testRunner.RemovePod(ctx, string(task.TestId))
							if err != nil {
								log.Errorf("cannot clean k8s resource %v %v", cleanK8sErr)
							}
						}
					}
				}
				//success state update by runner self
			}

			// startt init task
			initTasks, err := taskMgr.taskRepo.List(ctx, repo.ListParams{JobId: job.ID, State: []types.State{types.Init}})
			if err != nil {
				taskLog.Errorf("fetch task list fail %v", err)
				continue
			}

			for _, task := range initTasks {
				err = taskMgr.RunOneTask(ctx, task)
				if err != nil {
					taskLog.Errorf("fetch task list fail %v", err)
				}
			}
		}

		select {
		case <-ctx.Done():
			return nil
		case <-tm.C:
		}
	}
}

func (taskMgr *TaskMgr) RunOneTask(ctx context.Context, task *types.Task) error {
	pod, err := taskMgr.Process(ctx, task)
	if err != nil {
		markFailErr := taskMgr.taskRepo.MarkState(ctx, task.ID, types.Error, err.Error())
		if err != nil {
			return fmt.Errorf("cannot mark task as fail origin err %v %v", err, markFailErr)
		}
	}
	return taskMgr.taskRepo.UpdatePodRunning(ctx, task.ID, pod.Name)
}

func (taskMgr *TaskMgr) StopOneTask(ctx context.Context, id primitive.ObjectID) error {
	task, err := taskMgr.taskRepo.Get(ctx, id)
	if err != nil {
		return err
	}

	err = taskMgr.testRunner.CleanAll(ctx, string(task.TestId))
	if err != nil {
		return err
	}

	task.State = types.Error
	task.Logs = append(task.Logs, "stop manually")
	_, err = taskMgr.taskRepo.Save(ctx, task)
	if err != nil {
		return err
	}
	return nil
}

func (taskMgr *TaskMgr) Process(ctx context.Context, task *types.Task) (*corev1.Pod, error) {
	job, err := taskMgr.jobRepo.Get(ctx, task.JobId)
	if err != nil {
		return nil, err
	}

	testFlow, err := taskMgr.testFlowRepo.Get(ctx, &repo.GetTestFlowParams{ID: job.TestFlowId})
	if err != nil {
		taskLog.Errorf("get test flow failed %v", err)
		return nil, err
	}

	//confirm version and build image.
	taskLog.Infof("start to build image for testflow %s job %s", testFlow.Name, job.Name)
	versionMap, err := taskMgr.imageBuilder.BuildTestFlowEnv(ctx, testFlow.Nodes, job.Versions) //todo maybe move this code to previous step
	if err != nil {
		return nil, err
	}

	//save testflow as task params
	err = taskMgr.taskRepo.UpdateVersion(ctx, task.ID, versionMap)
	if err != nil {
		return nil, err
	}

	//run test flow
	file, err := os.Open(taskMgr.runnerConfig)
	if err != nil {
		return nil, err
	}

	return taskMgr.testRunner.ApplyRunner(ctx, file, map[string]string{
		"TaskID":          task.ID.Hex(),
		"TestId":          string(task.TestId),
		"PrivateRegistry": string(taskMgr.privateRegistry),
	})
}

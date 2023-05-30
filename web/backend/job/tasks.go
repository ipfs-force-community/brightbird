package job

import (
	"context"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/hunjixin/brightbird/models"

	errors2 "k8s.io/apimachinery/pkg/api/errors"

	"github.com/hunjixin/brightbird/repo"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/web/backend/config"
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
	cfg             config.Config
}

func NewTaskMgr(cfg config.Config, c *cron.Cron, jobRepo repo.IJobRepo, taskRepo repo.ITaskRepo, testFlowRepo repo.ITestFlowRepo, testRunner *TestRunnerDeployer, imageBuilder *ImageBuilderMgr, runnerConfig string, privateReg types.PrivateRegistry) *TaskMgr {
	return &TaskMgr{
		cfg:             cfg,
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
		taskLog.Infof("loop check task status and start new task")
		//scan tasks to process
		jobs, err := taskMgr.jobRepo.List(ctx) //todo 数据规模大了 可以考虑采用被动触发的方式 现在这种做法简单一些
		if err != nil {
			taskLog.Error("fetch job list fail %v", err)
			continue
		}
		for _, job := range jobs {
			//try to remove finish testrunner
			err := taskMgr.testRunner.RemoveFinishRunner(ctx)
			if err != nil {
				taskLog.Errorf("clean finish runner %v", err)
				continue
			}
			//check running task state
			runningTask, err := taskMgr.taskRepo.List(ctx, models.PageReq[repo.ListTaskParams]{
				PageNum:  1,
				PageSize: math.MaxInt64,
				Params: repo.ListTaskParams{
					JobID: job.ID,
					State: []models.State{models.Running, models.TempError},
				},
			})
			if err != nil {
				taskLog.Errorf("fetch running task list fail %v", err)
				continue
			}

			for _, task := range runningTask.List {
				restartCount, err := taskMgr.testRunner.CheckTestRunner(ctx, task.PodName)
				if err != nil {
					if errors2.IsNotFound(err) {
						markFailErr := taskMgr.taskRepo.MarkState(ctx, task.ID, models.Error, "not found testrunner, maybe delete manually")
						if markFailErr != nil {
							log.Errorf("cannot mark task as fail origin err %v %v", err, markFailErr)
						}
						continue
					} else {
						if restartCount > 5 {
							log.Errorf("task id(%s) name(%s) try exceed more than 5 times, mark error and remove", task.ID, task.Name)
							// mark pod as fail and remove this pod
							markFailErr := taskMgr.taskRepo.MarkState(ctx, task.ID, models.Error, "failed five times, delete task")
							if markFailErr != nil {
								log.Errorf("cannot mark task as fail %v origin err %v", err, markFailErr)
							}

							cleanK8sErr := taskMgr.testRunner.CleanTestResource(ctx, string(task.TestId))
							if cleanK8sErr != nil {
								log.Errorf("cannot clean k8s resource %v %v", cleanK8sErr)
							}
						}
					}
				}
				//success state update by runner self
			}

			// start init task
			initTasks, err := taskMgr.taskRepo.List(ctx, models.PageReq[repo.ListTaskParams]{
				PageNum:  1,
				PageSize: math.MaxInt64,
				Params: repo.ListTaskParams{
					JobID: job.ID,
					State: []models.State{models.Init},
				},
			})
			if err != nil {
				taskLog.Errorf("fetch task list fail %v", err)
				continue
			}

			for _, task := range initTasks.List {
				err = taskMgr.RunOneTask(ctx, task)
				if err != nil {
					taskLog.Errorf("run task list fail %v", err)
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

func (taskMgr *TaskMgr) RunOneTask(ctx context.Context, task *models.Task) error {
	pod, err := taskMgr.Process(ctx, task)
	if err != nil {
		markFailErr := taskMgr.taskRepo.MarkState(ctx, task.ID, models.Error, err.Error())
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

	err = taskMgr.testRunner.CleanTestResource(ctx, string(task.TestId))
	if err != nil {
		return err
	}

	task.State = models.Error
	task.Logs = append(task.Logs, "stop manually")
	_, err = taskMgr.taskRepo.Save(ctx, task)
	if err != nil {
		return err
	}
	return nil
}

func (taskMgr *TaskMgr) Process(ctx context.Context, task *models.Task) (*corev1.Pod, error) {
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
	commitMap, err := taskMgr.imageBuilder.BuildTestFlowEnv(ctx, testFlow.Nodes, task.InheritVersions) //todo maybe move this code to previous step
	if err != nil {
		return nil, err
	}

	//save testflow as task params
	err = taskMgr.taskRepo.UpdateCommitMap(ctx, task.ID, commitMap)
	if err != nil {
		return nil, err
	}

	//run test flow
	file, err := os.Open(taskMgr.runnerConfig)
	if err != nil {
		return nil, err
	}
	//--log-level=DEBUG, --namespace={{.NameSpace}},--config=/shared-dir/config-template.toml, --plugins=/shared-dir/plugins, --taskId={{.TaskID}}
	args := fmt.Sprintf(`"--logLevel=DEBUG", "--plugins=/shared-dir/plugins", "--tmpPath=/shared-dir/tmp", "--namespace=%s",  "--dbName=%s", "--mongoUrl=%s", "--mysql=%s", "--privReg=%s", "--taskId=%s"`,
		taskMgr.cfg.NameSpace,
		taskMgr.cfg.DbName,
		taskMgr.cfg.MongoUrl,
		taskMgr.cfg.Mysql,
		taskMgr.privateRegistry,
		task.ID.Hex())
	for _, p := range taskMgr.cfg.BootstrapPeers {
		args += fmt.Sprintf(` , "--bootPeer=%s" `, string(p))
	}
	return taskMgr.testRunner.ApplyRunner(ctx, file, map[string]string{
		"NameSpace":       taskMgr.cfg.NameSpace,
		"PrivateRegistry": string(taskMgr.privateRegistry),
		"TestId":          string(task.TestId),
		"Args":            args,
	})
}

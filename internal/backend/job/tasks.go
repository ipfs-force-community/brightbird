package job

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strconv"
	"sync"
	"time"

	"dario.cat/mergo"
	logging "github.com/ipfs/go-log/v2"
	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"

	"github.com/ipfs-force-community/brightbird/internal/backend/config"
	"github.com/ipfs-force-community/brightbird/pkg/env"
	"github.com/ipfs-force-community/brightbird/pkg/models"
	"github.com/ipfs-force-community/brightbird/pkg/repo"
	"github.com/ipfs-force-community/brightbird/pkg/types"
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
	tm := time.NewTicker(time.Second * 15)
	defer tm.Stop()
	taskLog.Info("loop check task status and start new task")

	for {
		//scan tasks to process
		jobs, err := taskMgr.jobRepo.List(ctx) //todo 数据规模大了 可以考虑采用被动触发的方式 现在这种做法简单一些
		if err != nil {
			taskLog.Error("fetch job list failed : %v", err)
			continue
		}
		wg := sync.WaitGroup{}
		for _, job := range jobs {
			//try to remove finish testRunner
			if err := taskMgr.testRunner.RemoveFinishRunner(ctx); err != nil {
				taskLog.Errorf("[testRunner] clean finish testRunner failed : %v", err)
				continue
			}
			wg.Add(1)
			go func(job *models.Job) {
				defer wg.Done()
				if err := taskMgr.doJob(ctx, job); err != nil {
					taskLog.With(
						"job", job.ID,
						"err", err,
					).Error("do job failed")
				}
			}(job)
		}
		wg.Wait()

		select {
		case <-ctx.Done():
			return nil
		case <-tm.C:
		}
	}
}

// doJob is a function to process test job
func (taskMgr *TaskMgr) doJob(ctx context.Context, job *models.Job) error {
	taskLog.With(
		"job", job.ID,
		"testFlowId", job.TestFlowId,
		"name", job.Name,
		"jobType", job.JobType,
		"execCount", job.ExecCount,
	).Debug("start to process job")

	taskMgr.doUpdateRunningTask(ctx, job)

	taskMgr.doUpdateInitTask(ctx, job)

	return nil
}

func (taskMgr *TaskMgr) doUpdateRunningTask(ctx context.Context, job *models.Job) error {
	// get running-task(Running State) list form jobId
	runningTask, err := taskMgr.taskRepo.List(ctx, models.PageReq[repo.ListTaskParams]{
		PageNum:  1,
		PageSize: math.MaxInt64,
		Params: repo.ListTaskParams{
			JobID: job.ID,
			State: []models.State{models.Running},
		},
	})
	if err != nil {
		taskLog.With(
			"job", job.ID,
			"state", []models.State{models.Running},
			"dbErr", err,
		).Errorf("fetch running task list failed")
		return err
	}

	// check test-runner state of running task
	for _, task := range runningTask.List {
		needClean := false
		if len(task.PodName) == 0 {
			//很少发生
			if dbErr := taskMgr.taskRepo.MarkState(ctx, task.ID, models.Error, "pod name is empty"); dbErr != nil {
				taskLog.With(
					"task", task.ID,
					"state-old", task.State,
					"state-new", models.Error,
					"msg", "pod name is empty",
					"dbErr", dbErr,
				).Errorf("update task state failed")
			}
			needClean = true
		}

		// checkout test-runner state from task.podName
		failedMsg := "failed 3 times, delete task"
		if _, err := taskMgr.testRunner.CheckTestRunner(ctx, task.PodName); err != nil {
			taskLog.With(
				"task", task.ID,
				"name", task.Name,
			).Errorf("try runner exceed more than 3 times, mark error and remove")
			// mark pod as fail and remove this pod
			if dbErr := taskMgr.taskRepo.MarkState(ctx, task.ID, models.Error, failedMsg); dbErr != nil {
				taskLog.With(
					"task", task.ID,
					"state-old", task.State,
					"state-new", models.Error,
					"msg", failedMsg,
					"dbErr", dbErr,
				).Errorf("update task state failed , origin error : %v", err)
			}
			needClean = true
		}

		// if test-runner has error, clean this test-runner
		if needClean {
			// clean test-runner rateled resource
			cleanK8sErr := taskMgr.testRunner.CleanTestResource(ctx, string(task.TestId)) //clean again to ensure release all resource
			if cleanK8sErr != nil {
				taskLog.With(
					"task", task.ID,
					"testId", task.TestId,
				).Errorf("clean k8s resource failed: %v", cleanK8sErr)
			}
		}
	}

	return nil
}

func (taskMgr *TaskMgr) doUpdateInitTask(ctx context.Context, job *models.Job) error {
	// start init task
	// get init-task(Building) list from jobId
	initTasks, err := taskMgr.taskRepo.List(ctx, models.PageReq[repo.ListTaskParams]{
		PageNum:  1,
		PageSize: math.MaxInt64,
		Params: repo.ListTaskParams{
			JobID: job.ID,
			State: []models.State{models.Init, models.Building},
		},
	})
	if err != nil {
		taskLog.With(
			"job", job.ID,
			"state", []models.State{models.Init, models.Building},
			"dbErr", err,
		).Errorf("fetch init&running task list failed")
		return err
	}
	var wg sync.WaitGroup
	for _, task := range initTasks.List {
		wg.Add(1)
		go func(task *models.Task) {
			defer wg.Done()
			// run task and return is test-runner pod
			pod, err := taskMgr.process(ctx, task)
			if err != nil {
				taskLog.With(
					"task", task.ID,
				).Errorf("process task fail : %v", err)

				if dbErr := taskMgr.taskRepo.MarkState(ctx, task.ID, models.Error, err.Error()); dbErr != nil {
					taskLog.With(
						"task", task.ID,
						"state-old", task.State,
						"state-new", models.Error,
						"msg", err.Error(),
						"dbErr", dbErr,
					).Errorf("update task state failed")
				}
				return
			}
			// update task-runner state to Running and update db
			if dbErr := taskMgr.taskRepo.UpdatePodRunning(ctx, task.ID, pod.Name); dbErr != nil {
				taskLog.With(
					"task", task.ID,
					"dbErr", dbErr,
				).Errorf("update task state running failed")
			}
		}(task)
	}
	wg.Wait()
	return nil
}
func (taskMgr *TaskMgr) StopOneTask(ctx context.Context, id primitive.ObjectID) error {
	task, err := taskMgr.taskRepo.Get(ctx, &repo.GetTaskReq{ID: id})
	if err != nil {
		return fmt.Errorf("db error [taskRepo.Get] : %v", err)
	}

	if err := taskMgr.testRunner.CleanTestResource(ctx, string(task.TestId)); err != nil {
		return fmt.Errorf("error CleanTestResource : %v", err)
	}

	task.State = models.Error
	task.Logs = append(task.Logs, "stop manually")

	if _, err := taskMgr.taskRepo.Save(ctx, task); err != nil {
		return fmt.Errorf("db error [taskRepo.Save] : %v", err)
	}
	return nil
}

func (taskMgr *TaskMgr) process(ctx context.Context, task *models.Task) (*corev1.Pod, error) {
	task, err := taskMgr.taskRepo.IncreaseRetry(ctx, task.ID)
	if err != nil {
		return nil, fmt.Errorf("db error [taskRepo.IncreaseRetry] : %v", err)
	}

	job, err := taskMgr.jobRepo.Get(ctx, task.JobId)
	if err != nil {
		return nil, err
	}

	testflow, err := taskMgr.testFlowRepo.Get(ctx, &repo.GetTestFlowParams{ID: job.TestFlowId})
	if err != nil {
		return nil, fmt.Errorf("db error [testFlowRepo.Get] : %v", err)
	}
	graph := &models.Graph{}
	err = yaml.Unmarshal([]byte(testflow.Graph), graph)
	if err != nil {
		return nil, err
	}

	// if task state is init
	if task.BeforeBuild() {
		// update task state to build completed
		if err := taskMgr.taskRepo.MarkState(ctx, task.ID, models.Building); err != nil {
			return nil, fmt.Errorf("db error [taskRepo.MarkState] : %v", err)
		}
		//todo 不在runner里面做的原因 1. 编译需要比较好的性能，而runner可能会调度到比较差的机器上 2. 网络问题，拉代码编译过程很慢， 代理太费流量， 这里使用同一份代码可以缓解一些

		//confirm version and build image.
		taskLog.With(
			"testflow", testflow.Name,
			"job", job.Name,
		).Infof("start build image")
		// create version map: key: node.name, value: version
		forceBuild := job.JobType == models.TagCreatedJobType
		commitMap, err := taskMgr.imageBuilder.BuildTestFlowEnv(ctx, graph.Pipeline, task.InheritVersions, forceBuild) //todo maybe move this code to previous step
		if err != nil {
			return nil, err
		}

		var pipelines []*types.ExecNode
		// construct a pipeline containing (input, output, instancename, name, plugintype, version) from graph.Pipeline
		for _, node := range graph.Pipeline {
			pipelines = append(pipelines, node.Value)
		}
		// save testflow as task params to db
		err = taskMgr.taskRepo.UpdatePipeline(ctx, task.ID, pipelines)
		if err != nil {
			return nil, fmt.Errorf("db error [taskRepo.UpdatePipeline] : %v", err)
		}

		// if user input version of plugin, it will be add into db as commitmap
		err = taskMgr.taskRepo.UpdateCommitMap(ctx, task.ID, commitMap)
		if err != nil {
			return nil, fmt.Errorf("db error [taskRepo.UpdateCommitMap]  : %v", err)
		}
	}

	//run test flow
	file, err := os.Open(taskMgr.runnerConfig)
	if err != nil {
		return nil, err
	}

	var defaultGlobal = make(env.GlobalParams)
	defaultGlobal["logLevel"] = "DEBUG"

	//append global config
	err = mergo.Merge(&defaultGlobal, env.GlobalParams(taskMgr.cfg.CustomProperties))
	if err != nil {
		return nil, err
	}

	//append testflow params
	for _, value := range testflow.GlobalProperties {
		defaultGlobal[value.Name] = value.Value
	}

	//append job global params
	for _, value := range job.GlobalProperties {
		defaultGlobal[value.Name] = value.Value
	}

	//yaml escape character
	globalParamsBytes, err := json.Marshal(defaultGlobal)
	if err != nil {
		return nil, err
	}

	globalParamsBytes, err = yaml.Marshal(string(globalParamsBytes))
	if err != nil {
		return nil, err
	}

	args := fmt.Sprintf(`"--plugins=%s", "--namespace=%s", "--retry=%d", "--dbName=%s", "--mongoUrl=%s", "--mysql=%s", "--registry=%s", "--taskId=%s", --globalParams, %s`,
		taskMgr.cfg.PluginStore,
		taskMgr.cfg.NameSpace,
		task.RetryTime,
		taskMgr.cfg.DBName,
		taskMgr.cfg.MongoURL,
		taskMgr.cfg.Mysql,
		taskMgr.privateRegistry,
		task.ID.Hex(),
		string(globalParamsBytes),
	)

	taskLog.Debugf("invoke testrunner args %s", args)

	// run test-runner and get test-runner pod
	return taskMgr.testRunner.ApplyRunner(ctx, file, map[string]string{
		"NameSpace": taskMgr.cfg.NameSpace,
		"Registry":  string(taskMgr.privateRegistry),
		"TestID":    string(task.TestId),
		"ReTry":     strconv.Itoa(task.RetryTime),
		"Args":      args,
	})
}

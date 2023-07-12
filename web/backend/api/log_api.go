package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hunjixin/brightbird/env/plugin"
	"github.com/hunjixin/brightbird/models"
	"github.com/hunjixin/brightbird/repo"
	ordered_map "github.com/wk8/go-ordered-map"
	"gopkg.in/yaml.v3"
)

func RegisterLogRouter(ctx context.Context, v1group *gin.RouterGroup, logRepo repo.ILogRepo, testflowRepo repo.ITestFlowRepo, taskRepo repo.ITaskRepo) {
	group := v1group.Group("/logs")

	// swagger:route GET /logs/pods/{testid} log listPodsInTest
	//
	// List all pod names in test.
	//
	//     Consumes:
	//
	//     Produces:
	//     - application/json
	//
	//     Schemes: http, https
	//
	//     Deprecated: false
	//
	//     Parameters:
	//       + name: testid
	//         in: path
	//         description: test id
	//         required: true
	//         type: string
	//
	//     Responses:
	//       200: stringArr
	//		 503: apiError
	group.GET("pods/:testid", func(c *gin.Context) {
		testID := c.Param("testid")
		pods, err := logRepo.ListPodsInTest(c, testID)
		if err != nil {
			c.Error(err) //nolint
			return
		}
		c.JSON(http.StatusOK, pods)
	})

	// swagger:route GET /logs/{podName} log podLogReq
	//
	// get all logs in pod.
	//
	//     Consumes:
	//
	//     Produces:
	//     - application/json
	//
	//     Schemes: http, https
	//
	//     Deprecated: false
	//
	//     Responses:
	//       200: logResp
	//		 503: apiError
	group.GET("logs", func(c *gin.Context) {
		podReq := models.PodLogReq{}
		err := c.ShouldBindQuery(&podReq)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		logs, err := logRepo.GetPodLog(c, podReq.PodName)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		resp := &models.LogResp{
			PodName: podReq.PodName,
			Logs:    logs,
		}

		if strings.Contains(podReq.PodName, "test-runner") {
			task, err := taskRepo.Get(ctx, &repo.GetTaskReq{TestId: &podReq.TestID})
			if err != nil {
				c.Error(err) //nolint
				return
			}

			steps, err := getRunnerLog(ctx, testflowRepo, task, logs)
			if err != nil {
				c.Error(err) //nolint
				return
			}
			resp.Steps = steps
		}
		c.JSON(http.StatusOK, resp)
	})
}

func getRunnerLog(ctx context.Context, testflowRepo repo.ITestFlowRepo, task *models.Task, logs []string) ([]*models.StepLog, error) {
	testflow, err := testflowRepo.Get(ctx, &repo.GetTestFlowParams{
		ID: task.TestFlowId,
	})
	if err != nil {
		return nil, err
	}
	graph := &models.Graph{}
	err = yaml.Unmarshal([]byte(testflow.Graph), graph)
	if err != nil {
		return nil, err
	}

	const preName = "testrunner prepare"
	const postName = "testrunner post"
	stepLogs := ordered_map.New()
	stepLogs.Set(preName, &models.StepLog{
		Name:         preName,
		InstanceName: preName,
		State:        models.StepNotRunning,
	})
	for _, pipe := range graph.Pipeline {
		stepLogs.Set(pipe.Key, &models.StepLog{
			Name:         pipe.Key,
			InstanceName: pipe.Value.Name,
			State:        models.StepNotRunning,
		})
	}

	stepLogs.Set(postName, &models.StepLog{
		Name:         postName,
		InstanceName: postName,
		State:        models.StepNotRunning,
	})

	var lines []string
	isFirst := true
	isClose := false
	var currentName string
	isRunnerCompleted := false
	for _, log := range logs {
		if !isRunnerCompleted {
			isRunnerCompleted = log == "RUNNEREND"
		}

		cmd, val, isCmd := plugin.ReadCMD(log)
		if isCmd {
			switch cmd {
			case plugin.CMDSTARTPREFIX:
				if isFirst {
					stepI, ok := stepLogs.Get(preName)
					if !ok {
						return nil, fmt.Errorf("%s not found", preName)
					}
					step := stepI.(*models.StepLog)
					step.Logs = lines
					step.State = models.StepSuccess
					isFirst = false
				}
				isClose = false
				currentName = val
				//reset
				lines = []string{}
			case plugin.CMDSUCCESSPREFIX:
				stepI, ok := stepLogs.Get(currentName)
				if !ok {
					return nil, fmt.Errorf("%s not found", currentName)
				}
				step := stepI.(*models.StepLog)
				step.Logs = append(lines, log)
				step.State = models.StepSuccess
				//reset
				lines = []string{}
				isClose = true
				continue
			case plugin.CMDERRORREFIX:
				stepI, ok := stepLogs.Get(currentName)
				if !ok {
					return nil, fmt.Errorf("%s not found", currentName)
				}
				step := stepI.(*models.StepLog)
				step.Logs = append(lines, log)
				step.State = models.StepFail

				isClose = true
				//reset
				lines = []string{}
				continue
			default:
			}
		}

		lines = append(lines, log)
	}

	if isClose {
		//not complete
		stepI, ok := stepLogs.Get(postName)
		if !ok {
			return nil, fmt.Errorf("%s not found", postName)
		}
		step := stepI.(*models.StepLog)
		step.Logs = lines
		if isRunnerCompleted {
			step.State = models.StepSuccess
		} else {
			step.State = models.StepRunning
		}
	} else {
		stepI, ok := stepLogs.Get(currentName)
		if !ok {
			return nil, fmt.Errorf("%s not found", currentName)
		}
		step := stepI.(*models.StepLog)
		step.State = models.StepRunning
		step.Logs = lines
	}

	var logsArray = []*models.StepLog{}
	cur := stepLogs.Oldest()
	for {
		logsArray = append(logsArray, cur.Value.(*models.StepLog))
		val := cur.Next()
		if val == nil {
			break
		}
		cur = val
	}
	return logsArray, nil
}

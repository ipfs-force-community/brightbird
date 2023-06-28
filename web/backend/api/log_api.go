package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hunjixin/brightbird/env/plugin"
	"github.com/hunjixin/brightbird/models"
	"github.com/hunjixin/brightbird/repo"
)

func RegisterLogRouter(ctx context.Context, v1group *gin.RouterGroup, logRepo repo.ILogRepo) {
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

	// swagger:route GET /logs/{podName} log listLogsInPod
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
	//     Parameters:
	//       + name: podName
	//         in: path
	//         description: pod name
	//         required: true
	//         type: string
	//
	//     Responses:
	//       200: logResp
	//		 503: apiError
	group.GET(":podName", func(c *gin.Context) {
		podName := c.Param("podName")
		logs, err := logRepo.GetPodLog(c, podName)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		resp := &models.LogResp{
			PodName: podName,
			Logs:    logs,
		}

		if strings.Contains(podName, "test-runner") {
			var stepLogs []models.StepLog
			var lines []string
			isClose := false
			currentSec := "testrunner prepare"
			for _, log := range logs {
				cmd, val, isCmd := plugin.ReadCMD(log)
				if isCmd {
					switch cmd {
					case plugin.CMDSTARTPREFIX:
						if isClose {
							isClose = false
						} else {
							stepLogs = append(stepLogs, models.StepLog{
								Name:      currentSec,
								IsSuccess: true,
								Logs:      lines,
							})
						}
						lines = []string{log}
						currentSec = val
						continue
					case plugin.CMDSUCCESSPREFIX:
						stepLogs = append(stepLogs, models.StepLog{
							Name:      currentSec,
							IsSuccess: true,
							Logs:      append(lines, log),
						})
						lines = []string{}
						isClose = true
						continue
					case plugin.CMDERRORREFIX:
						stepLogs = append(stepLogs, models.StepLog{
							Name:      currentSec,
							IsSuccess: false,
							Logs:      append(lines, log),
						})
						lines = []string{}
						isClose = true
						continue
					}
				}

				lines = append(lines, log)
			}

			if isClose {
				stepLogs = append(stepLogs, models.StepLog{
					Name:      "testrunner end",
					IsSuccess: true,
					Logs:      lines,
				})
				resp.Steps = stepLogs
			} else {
				//not complete
				stepLogs = append(stepLogs, models.StepLog{
					Name:      currentSec,
					IsSuccess: true,
					Logs:      lines,
				})
				resp.Steps = stepLogs
			}
		}
		c.JSON(http.StatusOK, resp)
	})
}

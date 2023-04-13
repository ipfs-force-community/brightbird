package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hunjixin/brightbird/repo"
)

// PodListResp
// swagger:model podListResp
type PodListResp []string

// LogListResp
// swagger:model logListResp
type LogListResp []string

func RegisterLogRouter(ctx context.Context, v1group *gin.RouterGroup, logRepo repo.ILogRepo) {
	group := v1group.Group("/logs")

	// swagger:route GET /logs/pods/{testid} listPodsInTest
	//
	// List all pod names in test.
	//
	//     Consumes:
	//     - application/json
	//
	//     Produces:
	//     - application/json
	//     - application/text
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
	//       200: podListResp
	group.GET("pods/:testid", func(c *gin.Context) {
		testID := c.Param("testid")
		pods, err := logRepo.ListPodsInTest(c, testID)
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(http.StatusOK, pods)
	})

	// swagger:route GET /logs/{podName} listLogsInPod
	//
	// get all logs in pod.
	//
	//     Consumes:
	//     - application/json
	//
	//     Produces:
	//     - application/json
	//     - application/text
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
	//       200: podListResp
	group.GET(":podName", func(c *gin.Context) {
		podName := c.Param("podName")
		pods, err := logRepo.GetPodLog(c, podName)
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(http.StatusOK, pods)
	})
}

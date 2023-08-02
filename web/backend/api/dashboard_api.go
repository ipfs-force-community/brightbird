package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ipfs-force-community/brightbird/models"
	"github.com/ipfs-force-community/brightbird/repo"
	"github.com/ipfs-force-community/brightbird/types"
	"go.mongodb.org/mongo-driver/mongo"
)

func RegisterDashboardRouter(ctx context.Context, v1group *V1RouterGroup, tasksRepo repo.ITaskRepo,
	pluginRepo repo.IPluginService, jobRepo repo.IJobRepo) {
	group := v1group.Group("/dashboard")

	// swagger:route GET /task-count tasks getTaskCount
	//
	// Retrieves the Statistics of tasks.
	//
	//     Produces:
	//     - application/json
	//
	//     Schemes: http, https
	//
	//
	//     Responses:
	//       200: myString   //todo fix correctstruct
	//       500: apiError
	group.GET("/task-count", func(c *gin.Context) {
		total, err := tasksRepo.CountAllAmount(ctx)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		passedAmount, err := tasksRepo.CountAllAmount(ctx, models.Successful)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		failedAmount, err := tasksRepo.CountAllAmount(ctx, models.Error)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		passRate := float64(passedAmount) / float64(total) * 100

		passRateFormatted := fmt.Sprintf("%.1f%%", passRate)

		response := map[string]interface{}{
			"total":    total,
			"passed":   passedAmount,
			"failed":   failedAmount,
			"passRate": passRateFormatted,
		}

		c.JSON(http.StatusOK, response)
	})

	// swagger:route GET /test-data test-data listTestData
	//
	// Lists test data.
	//
	// Lists the amount of tasks for a job in the last 2 weeks.
	//
	//     Produces:
	//     - application/json
	//
	//     Schemes: http
	//
	//
	//     Responses:
	//       200: myString   //todo fix correctstruct
	//       500: apiError
	group.GET("/test-data", func(c *gin.Context) {
		testData, dateArray, err := tasksRepo.TaskAmountOfJobLast2Week(ctx)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		newData := make(map[string][]int)
		for jobId, countArray := range testData {
			job, err := jobRepo.Get(ctx, jobId)
			if err != nil {
				if err == mongo.ErrNoDocuments {
					continue
				}
				c.Error(err) //nolint
				return
			}
			newData[job.Name] = countArray
		}

		response := map[string]interface{}{
			"testData":  newData,
			"dateArray": dateArray,
		}

		c.JSON(http.StatusOK, response)
	})

	// swagger:route GET /today-pass-rate-ranking task todayPassRateRankingReq
	//
	// Retrieves the top 3 job pass rates for today.
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
	//
	//     Responses:
	//       200: myString   //todo fix correctstruct
	//       500: apiError
	group.GET("/today-pass-rate-ranking", func(c *gin.Context) {
		jobIds, passRates, err := tasksRepo.JobPassRateTop3Today(ctx)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		jobNames := []string{}

		for _, jobId := range jobIds {
			job, err := jobRepo.Get(ctx, jobId)
			if err != nil {
				c.Error(err) //nolint
				return
			}

			jobNames = append(jobNames, job.Name)
		}

		response := map[string]interface{}{
			"jobNames":  jobNames,
			"passRates": passRates,
		}

		c.JSON(http.StatusOK, response)
	})

	// swagger:route GET /failed-tasks failed-tasks listFailedTasksReq
	//
	// Lists the failed tasks.
	//
	//     Produces:
	//     - application/json
	//
	//     Schemes: http, https
	//
	//     Responses:
	//       200: myString   //todo fix correctstruct
	//       500: apiError
	group.GET("/failed-tasks", func(c *gin.Context) {
		failTask, err := tasksRepo.JobFailureRatiobLast2Week(ctx)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		failTaskWithJobname := make(map[string]interface{})
		for jobId, value := range failTask {
			job, err := jobRepo.Get(ctx, jobId)
			if err != nil {
				if err == mongo.ErrNoDocuments {
					continue
				}
				c.Error(err) //nolint
				return
			}
			failTaskWithJobname[job.Name] = value
		}

		response := map[string]interface{}{
			"failTask": failTaskWithJobname,
		}

		c.JSON(http.StatusOK, response)
	})

	// swagger:route GET /pass-rate-trends task passRateTrends
	//
	// Gets the pass rate trends for the last 30 days.
	//
	//     Produces:
	//     - application/json
	//
	//     Schemes: http, https
	//
	//     Responses:
	//       200: myString   //todo fix correctstruct
	//       500: apiError
	group.GET("/pass-rate-trends", func(c *gin.Context) {
		dateArray, passRateArray, err := tasksRepo.TasktPassRateLast30Days(ctx)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		response := map[string]interface{}{
			"dateArray":     dateArray,
			"passRateArray": passRateArray,
		}

		c.JSON(http.StatusOK, response)
	})

	// swagger:route GET /count-plugins plugin countPluginsReq
	//
	// Counts the number of plugins.
	//
	//     Produces:
	//     - application/json
	//
	//     Schemes: http, https
	//
	//     Deprecated: false
	//
	//
	//     Responses:
	//       200: myString   //todo fix correctstruct
	//       500: apiError
	group.GET("/count-plugins", func(c *gin.Context) {
		deployerPluginType := types.Deploy
		deployerCount, err := pluginRepo.CountPlugin(ctx, &deployerPluginType)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		execPluginType := types.TestExec
		execCount, err := pluginRepo.CountPlugin(ctx, &execPluginType)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		response := map[string]interface{}{
			"deployerCount": deployerCount,
			"execCount":     execCount,
		}

		c.JSON(http.StatusOK, response)

	})

	// swagger:route GET /success-quantity-trends success-quantity-trends successQuantityTrendsReq
	//
	// Retrieves the success quantity trends.
	//
	//     Produces:
	//     - application/json
	//
	//     Schemes: http, https
	//
	//     Responses:
	//       200: myString   //todo fix correctstruct
	//       500: apiError
	group.GET("/success-quantity-trends", func(c *gin.Context) {
		testData, dateArray, err := tasksRepo.JobPassRateLast30Days(ctx)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		newData := make(map[string][]int)
		for jobId, countArray := range testData {
			job, err := jobRepo.Get(ctx, jobId)
			if err != nil {
				if err == mongo.ErrNoDocuments {
					continue
				}
				c.Error(err) //nolint
				return
			}
			newData[job.Name] = countArray
		}

		response := map[string]interface{}{
			"testData":  newData,
			"dateArray": dateArray,
		}

		c.JSON(http.StatusOK, response)
	})
}

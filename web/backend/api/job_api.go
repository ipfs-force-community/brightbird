package api

import (
	"context"
	"math"
	"net/http"

	"github.com/hunjixin/brightbird/models"

	"github.com/gin-gonic/gin"
	"github.com/hunjixin/brightbird/repo"
	"github.com/hunjixin/brightbird/web/backend/job"
	logging "github.com/ipfs/go-log/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var jobLogger = logging.Logger("job_api")

func RegisterJobRouter(ctx context.Context, v1group *V1RouterGroup, jobRepo repo.IJobRepo, taskRepo repo.ITaskRepo, testFlowRepo repo.ITestFlowRepo, groupRepo repo.IGroupRepo, jobManager job.IJobManager, taskManager *job.TaskMgr) {
	group := v1group.Group("/job")
	// swagger:route GET /job/list job listJobs
	//
	// Lists all jobs.
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
	//     Responses:
	//       200: listJobResp
	//		 503: apiError
	group.GET("list", func(c *gin.Context) {
		jobs, err := jobRepo.List(ctx)
		if err != nil {
			c.Error(err) //nolint
			return
		}
		c.JSON(http.StatusOK, jobs)
	})

	// swagger:route GET /job/count job countJobRequest
	//
	// Count all jobs by condition.
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
	//     Responses:
	//       200: listJobResp
	//		 503: apiError
	group.GET("count", func(c *gin.Context) {
		req := &models.CountJobRequest{}
		err := c.ShouldBindQuery(req)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		params := &repo.CountJobParams{
			Name: req.Name,
		}

		if req.ID != nil {
			params.ID, err = primitive.ObjectIDFromHex(*req.ID)
			if err != nil {
				c.Error(err) //nolint
				return
			}
		}

		count, err := jobRepo.Count(ctx, params)
		if err != nil {
			c.Error(err) //nolint
			return
		}
		c.JSON(http.StatusOK, count)
	})

	// swagger:route Get /job/{id} job getJob
	//
	// Get job by id
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
	//       + name: id
	//         in: path
	//         description: job id
	//         required: true
	//         type: string
	//
	//     Responses:
	//       200: job
	//		 503: apiError
	group.GET(":id", func(c *gin.Context) {
		id, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.Error(err) //nolint
			return
		}

		job, err := jobRepo.Get(ctx, id)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		c.JSON(http.StatusOK, job)
	})

	// swagger:route Get /job/detail/{id} job getJob
	//
	// Get job detail by id
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
	//       + name: id
	//         in: path
	//         description: job id
	//         required: true
	//         type: string
	//
	//     Responses:
	//       200: jobDetailResp
	//		 503: apiError
	group.GET("detail/:id", func(c *gin.Context) {
		id, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.Error(err) //nolint
			return
		}

		job, err := jobRepo.Get(ctx, id)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		testflow, err := testFlowRepo.Get(ctx, &repo.GetTestFlowParams{ID: job.TestFlowId})
		if err != nil {
			c.Error(err) //nolint
			return
		}

		tfGroup, err := groupRepo.Get(ctx, testflow.GroupId)
		if err != nil {
			c.Error(err) //nolint
			return
		}
		c.JSON(http.StatusOK, models.JobDetailResp{
			Job:          *job,
			TestFlowName: testflow.Name,
			GroupName:    tfGroup.Name,
		})
	})

	// swagger:route Get /job/{id} job updateJob
	//
	// Update job
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
	//       + name: id
	//         in: path
	//         description: job id
	//         required: true
	//         type: string
	//       + name: updateJobParams
	//         in: body
	//         description: job update params
	//         required: true
	//         type: updateJobRequest
	//
	//     Responses:
	//       200: job
	//		 503: apiError
	group.POST(":id", func(c *gin.Context) {
		params := &models.UpdateJobRequest{}
		err := c.ShouldBindJSON(params)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		id, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.Error(err) //nolint
			return
		}

		job, err := jobRepo.Get(ctx, id)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		job.TestFlowId = params.TestFlowId
		job.Name = params.Name
		job.Description = params.Description
		job.CronJobParams = params.CronJobParams
		job.Versions = params.Versions

		err = job.CheckParams()
		if err != nil {
			c.Error(err) //nolint
			return
		}

		_, err = jobRepo.Save(ctx, job)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		err = jobManager.InsertOrReplaceJob(ctx, job)
		if err != nil {
			c.Error(err) //nolint
			return
		}
		c.JSON(http.StatusOK, job)
	})

	// swagger:route DELETE /job/{id} job deleteJob
	//
	// Delete job by id
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
	//       + name: id
	//         in: path
	//         description: id of  job
	//         required: true
	//         type: string
	//
	//     Responses:
	//       200:
	//		 503: apiError
	group.DELETE("/:id", func(c *gin.Context) {
		id, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.Error(err) //nolint
			return
		}

		err = jobRepo.Delete(c, id)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		//remove job
		err = jobManager.StopJob(ctx, id)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		//remove task
		tasks, err := taskRepo.List(ctx, models.PageReq[repo.ListTaskParams]{
			PageNum:  1,
			PageSize: math.MaxInt64,
			Params: repo.ListTaskParams{
				JobID: id,
			},
		})
		if err != nil {
			c.Error(err) //nolint
			return
		}

		for _, task := range tasks.List {
			if task.State == models.Running || task.State == models.TempError {
				err = taskManager.StopOneTask(ctx, task.ID)
				if err != nil {
					jobLogger.Warnf("delete job, but clean task fail and need clean manually %s", err)
				}
				err = taskRepo.Delete(ctx, task.ID)
				if err != nil {
					jobLogger.Warnf("delete task fail %v", err)
				}
			}
		}

		err = taskRepo.Delete(ctx, id)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		c.Status(http.StatusOK)
	})

	// swagger:route POST /job job saveJob
	//
	// save job entity, create if not exist
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
	//       + name: job
	//         in: body
	//         description: job json
	//         required: true
	//         type: job
	//         allowEmpty:  false
	//
	//     Responses:
	//       200:
	//		 503: apiError
	group.POST("", func(c *gin.Context) {
		job := &models.Job{}
		err := c.ShouldBindJSON(job)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		err = job.CheckParams()
		if err != nil {
			c.Error(err) //nolint
			return
		}

		id, err := jobRepo.Save(ctx, job)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		err = jobManager.InsertOrReplaceJob(ctx, job)
		if err != nil {
			c.Error(err) //nolint
			return
		}
		c.String(http.StatusOK, id.Hex())
	})

	// swagger:route POST /run/{jobid} job runJobImmediately
	// run job immediately
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
	//       + name: jobid
	//         in: path
	//         description: job id
	//         required: true
	//         type: string
	//         allowEmpty:  false
	//
	//     Responses:
	//       200:
	//		 503: apiError
	group.POST("/run/:jobid", func(c *gin.Context) {
		jobIDStr := c.Param("jobid")
		jobId, err := primitive.ObjectIDFromHex(jobIDStr)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		taskID, err := jobManager.ExecJobImmediately(c, jobId)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		c.String(http.StatusOK, taskID.Hex())
	})
}

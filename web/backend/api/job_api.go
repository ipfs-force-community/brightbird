package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hunjixin/brightbird/repo"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/web/backend/job"
	logging "github.com/ipfs/go-log/v2"
	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var jobLogger = logging.Logger("job_api")

// UpdateJobRequest
// swagger:model updateJobRequest
type UpdateJobRequest struct {
	TestFlowId  primitive.ObjectID `json:"testFlowId"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	//cron job params
	types.CronJobParams
	Versions map[string]string `json:"versions"`
}

// JobDetailResp
// swagger:model jobDetailResp
type JobDetailResp struct {
	types.Job
	TestFlowName string `json:"testFlowName"`
	GroupName    string `json:"groupName"`
}

// ListJobResp
// swagger:model listJobResp
type ListJobResp []types.Job

func RegisterJobRouter(ctx context.Context, v1group *V1RouterGroup, jobRepo repo.IJobRepo, taskRepo repo.ITaskRepo, testFlowRepo repo.ITestFlowRepo, groupRepo repo.IGroupRepo, jobManager job.IJobManager, taskManager *job.TaskMgr) {
	group := v1group.Group("/job")

	// swagger:route GET /job listJobs
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
	group.GET("list", func(c *gin.Context) {
		jobs, err := jobRepo.List(ctx)
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(http.StatusOK, jobs)
	})

	// swagger:route Get /job/{id} getJob
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
	group.GET(":id", func(c *gin.Context) {
		id, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.Error(err)
			return
		}

		job, err := jobRepo.Get(ctx, id)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, job)
	})

	// swagger:route Get /job/detail/{id} getJob
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
	group.GET("detail/:id", func(c *gin.Context) {
		id, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.Error(err)
			return
		}

		job, err := jobRepo.Get(ctx, id)
		if err != nil {
			c.Error(err)
			return
		}

		testflow, err := testFlowRepo.Get(ctx, &repo.GetTestFlowParams{ID: job.TestFlowId})
		if err != nil {
			c.Error(err)
			return
		}

		tfGroup, err := groupRepo.Get(ctx, testflow.GroupId)
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(http.StatusOK, JobDetailResp{
			*job,
			testflow.Name,
			tfGroup.Name,
		})
	})

	// swagger:route Get /job/{id} updateJob
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
	group.POST(":id", func(c *gin.Context) {
		params := &UpdateJobRequest{}
		err := c.ShouldBindJSON(params)
		if err != nil {
			c.Error(err)
			return
		}

		id, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.Error(err)
			return
		}

		job, err := jobRepo.Get(ctx, id)
		if err != nil {
			c.Error(err)
			return
		}

		job.TestFlowId = params.TestFlowId
		job.Name = params.Name
		job.Description = params.Description
		job.CronJobParams = params.CronJobParams
		job.Versions = params.Versions

		switch job.JobType {
		case types.CronJobType:
			_, err = cron.ParseStandard(job.CronExpression)
			if err != nil {
				c.Error(err)
				return
			}
		}

		_, err = jobRepo.Save(ctx, job)
		if err != nil {
			c.Error(err)
			return
		}

		err = jobManager.InsertOrReplaceJob(ctx, job)
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(http.StatusOK, job)
	})

	// swagger:route DELETE /job/{id} deleteJob
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
	group.DELETE("/:id", func(c *gin.Context) {
		id, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.Error(err)
			return
		}

		err = jobRepo.Delete(c, id)
		if err != nil {
			c.Error(err)
			return
		}

		//remove job
		err = jobManager.StopJob(ctx, id)
		if err != nil {
			c.Error(err)
			return
		}

		/*
			//remove task
			tasks, err := taskRepo.List(ctx, repo.ListParams{
				JobId: id,
			})
			if err != nil {
				c.Error(err)
				return
			}

			for _, task := range tasks {
				err = taskManager.StopOneTask(ctx, task.ID)
				if err != nil {
					jobLogger.Warnf("delete job, but clean task fail and need clean manually %s", err)
				}
			}

			err = taskRepo.DeleteByJobId(ctx, id)
			if err != nil {
				c.Error(err)
				return
			}
		*/
		c.Status(http.StatusOK)
	})

	// swagger:route POST /job saveJob
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
	group.POST("", func(c *gin.Context) {
		job := &types.Job{}
		err := c.ShouldBindJSON(job)
		if err != nil {
			c.Error(err)
			return
		}

		switch job.JobType {
		case types.CronJobType:
			_, err = cron.ParseStandard(job.CronExpression)
			if err != nil {
				c.Error(err)
				return
			}
		}

		id, err := jobRepo.Save(ctx, job)
		if err != nil {
			c.Error(err)
			return
		}

		err = jobManager.InsertOrReplaceJob(ctx, job)
		if err != nil {
			c.Error(err)
			return
		}
		c.String(http.StatusOK, id.Hex())
	})
}

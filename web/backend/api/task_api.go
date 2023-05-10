package api

import (
	"context"
	"github.com/hunjixin/brightbird/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hunjixin/brightbird/repo"
	"github.com/hunjixin/brightbird/web/backend/job"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func RegisterTaskRouter(ctx context.Context, v1group *V1RouterGroup, taskManager *job.TaskMgr, tasksRepo repo.ITaskRepo) {
	group := v1group.Group("/task")

	// swagger:route GET /task listTasks
	//
	// Lists all tasks.
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
	//       + name: jobId
	//         in: query
	//         description: job id
	//         required: false
	//         type: string
	//
	//     Responses:
	//       200: listTasksResp
	//		 503: apiError
	group.GET("list", func(c *gin.Context) {
		params := models.ListTasksReq{}
		err := c.ShouldBindWith(&params, paginationQueryBind)
		if err != nil {
			c.Error(err)
			return
		}

		jobId, err := primitive.ObjectIDFromHex(params.Params.JobId)
		if err != nil {
			c.Error(err)
			return
		}

		tasks, err := tasksRepo.List(ctx, models.PageReq[repo.ListTaskParams]{
			PageNum:  params.PageNum,
			PageSize: params.PageSize,
			Params: repo.ListTaskParams{
				JobId: jobId,
				State: params.Params.State,
			},
		})
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(http.StatusOK, tasks)
	})

	// swagger:route Get /task/{id} getTask
	//
	// Get task by id
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
	//       200: task
	//		 503: apiError
	group.GET(":id", func(c *gin.Context) {
		id, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.Error(err)
			return
		}

		task, err := tasksRepo.Get(ctx, id)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, task)
	})

	// swagger:route DELETE /task/{id} deleteTask
	//
	// Delete task by id
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
			c.Error(err)
			return
		}
		err = tasksRepo.Delete(c, id)
		if err != nil {
			c.Error(err)
			return
		}

		c.Status(http.StatusOK)
	})

	// swagger:route DELETE /task/stop/{id} stopTask
	//
	// stop task
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
	group.POST("/stop/:id", func(c *gin.Context) {
		id, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.Error(err)
			return
		}
		taskManager.StopOneTask(ctx, id)
		c.Status(http.StatusOK)
	})
}

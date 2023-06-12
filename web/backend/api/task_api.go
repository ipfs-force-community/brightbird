package api

import (
	"context"
	"net/http"

	"github.com/hunjixin/brightbird/models"

	"github.com/gin-gonic/gin"
	"github.com/hunjixin/brightbird/repo"
	"github.com/hunjixin/brightbird/web/backend/job"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func RegisterTaskRouter(ctx context.Context, v1group *V1RouterGroup, taskManager *job.TaskMgr, tasksRepo repo.ITaskRepo) {
	group := v1group.Group("/task")

	// swagger:route GET /task task listTasksReq
	//
	// Lists all tasks.
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
	//       200: listTasksResp
	//		 503: apiError
	group.GET("list", func(c *gin.Context) {
		params := models.ListTasksReq{}
		err := c.ShouldBindQuery(&params)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		JobID, err := primitive.ObjectIDFromHex(params.JobID)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		tasks, err := tasksRepo.List(ctx, models.PageReq[repo.ListTaskParams]{
			PageNum:  params.PageNum,
			PageSize: params.PageSize,
			Params: repo.ListTaskParams{
				JobID:      JobID,
				State:      params.State,
				CreateTime: params.CreateTime,
			},
		})
		if err != nil {
			c.Error(err) //nolint
			return
		}
		c.JSON(http.StatusOK, tasks)
	})

	// swagger:route Get /task/{id} task getTask
	//
	// Get task by id
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
			c.Error(err) //nolint
			return
		}

		task, err := tasksRepo.Get(ctx, id)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		c.JSON(http.StatusOK, task)
	})

	// swagger:route DELETE /task/{id} task deleteTask
	//
	// Delete task by id
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
		err = tasksRepo.Delete(c, id)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		c.Status(http.StatusOK)
	})

	// swagger:route DELETE /task/stop/{id} task stopTask
	//
	// stop task
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
			c.Error(err) //nolint
			return
		}
		err = taskManager.StopOneTask(ctx, id)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		c.Status(http.StatusOK)
	})
}

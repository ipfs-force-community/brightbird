package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ipfs-force-community/brightbird/models"

	"github.com/gin-gonic/gin"
	"github.com/ipfs-force-community/brightbird/repo"
	"github.com/ipfs-force-community/brightbird/web/backend/job"
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

	// swagger:route Get /task task getTaskReq
	//
	// Get task by condition
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
	//       200: task
	//		 503: apiError
	group.GET("", func(c *gin.Context) {
		var req models.GetTaskReq
		err := c.ShouldBindQuery(&req)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		req2 := &repo.GetTaskReq{TestID: req.TestID}
		if req.ID != nil {
			req2.ID, err = primitive.ObjectIDFromHex(*req.ID)
			if err != nil {
				c.Error(err) //nolint
				return
			}
		}

		task, err := tasksRepo.Get(ctx, req2)
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

	// swagger:route Get /task/retry task retryTaskReq
	//
	// Retry a fail task
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
	//       200: int64Arr
	//		 503: apiError
	group.POST("retry", func(c *gin.Context) {
		retryTaskReq := &models.RetryTaskReq{}
		err := c.ShouldBindJSON(retryTaskReq)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		taskID, err := primitive.ObjectIDFromHex(retryTaskReq.ID)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		task, err := tasksRepo.Get(ctx, &repo.GetTaskReq{ID: taskID})
		if err != nil {
			c.Error(err) //nolint
			return
		}

		if task.State != models.Error {
			c.Error(fmt.Errorf("only retry error task")) //nolint
			return
		}

		err = tasksRepo.MarkState(ctx, taskID, models.Init, "task retry")
		if err != nil {
			c.Error(err) //nolint
			return
		}
		c.Status(http.StatusOK)
	})
}

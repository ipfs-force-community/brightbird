package api

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/hunjixin/brightbird/repo"
	"github.com/hunjixin/brightbird/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
)

// ListTaskResp
// swagger:model listTaskResp
type ListTaskResp []types.Task

func RegisterTaskAPI(ctx context.Context, v1group *V1RouterGroup, tasksRepo repo.ITaskRepo) {
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
	//     Responses:
	//       200: listTaskResp
	group.GET("", func(c *gin.Context) {
		tasks, err := tasksRepo.List(ctx)
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(http.StatusOK, tasks)
	})

	// swagger:route Get /task getTask
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
	group.GET(":id", func(c *gin.Context) {
		id, err := primitive.ObjectIDFromHex(c.Query("id"))
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

	// swagger:route Get /task listTasksInJob
	//
	// Get tasks belong to specific job
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
	//         required: true
	//         type: string
	//
	//     Responses:
	//       200: listTaskResp
	group.GET("", func(c *gin.Context) {
		jobId, err := primitive.ObjectIDFromHex(c.Query("jobId"))
		if err != nil {
			c.Error(err)
			return
		}

		tasks, err := tasksRepo.ListInJob(ctx, jobId)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, tasks)
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
}

package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hunjixin/brightbird/repo"
	"github.com/hunjixin/brightbird/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ListTaskResp
// swagger:model listTaskResp
type ListTaskResp []types.Task

func RegisterTaskRouter(ctx context.Context, v1group *V1RouterGroup, tasksRepo repo.ITaskRepo) {
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
	//       200: listTaskResp
	group.GET("", func(c *gin.Context) {
		params := repo.ListParams{}
		err := c.ShouldBindQuery(&params)
		if err != nil {
			c.Error(err)
			return
		}

		tasks, err := tasksRepo.List(ctx, params)
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

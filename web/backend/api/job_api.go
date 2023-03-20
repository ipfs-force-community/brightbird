package api

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/hunjixin/brightbird/repo"
	"github.com/hunjixin/brightbird/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
)

// ListJobResp
// swagger:model listJobResp
type ListJobResp []types.Job

func RegisterJobRouter(ctx context.Context, v1group *V1RouterGroup, jobRepo repo.IJobRepo) {
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
	group.GET("", func(c *gin.Context) {
		jobs, err := jobRepo.List(ctx)
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(http.StatusOK, jobs)
	})

	// swagger:route Get /job getJob
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
		id, err := primitive.ObjectIDFromHex(c.Query("id"))
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

		c.Status(http.StatusOK)
	})
}

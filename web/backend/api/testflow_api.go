package api

import (
	"context"
	"github.com/hunjixin/brightbird/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hunjixin/brightbird/repo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func RegisterTestFlowRouter(ctx context.Context, v1group *V1RouterGroup, service repo.ITestFlowRepo) {
	group := v1group.Group("/testflow")
	// swagger:route GET /testflow/list/ listTestFlowsInGroup
	//
	// Lists exec test flows in specific group.
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
	//       + name: groupId
	//         in: query
	//         description: group id  of test flow
	//         required: true
	//         type: string
	//       + name: pageNum
	//         in: query
	//         description: page number  of test flow
	//         required: false
	//         type: integer
	//       + name: pageSize
	//         in: query
	//         description: page size  of test flow
	//         required: false
	//         type: integer
	//
	//     Responses:
	//       200: listTestFlowResp
	//		 503: apiError
	group.GET("list", func(c *gin.Context) {
		req := &models.ListInGroupRequest{}
		err := c.ShouldBindWith(req, paginationQueryBind)
		if err != nil {
			c.Error(err)
			return
		}

		groupId, err := primitive.ObjectIDFromHex(req.Params.GroupId)
		if err != nil {
			c.Error(err)
			return
		}

		output, err := service.List(ctx, models.PageReq[repo.ListTestFlowParams]{
			PageNum:  req.PageNum,
			PageSize: req.PageSize,
			Params: repo.ListTestFlowParams{
				GroupID: groupId,
				Name:    req.Params.Name,
			},
		})
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, output)
	})

	// swagger:route GET /testflow/count/ countTestFlowsInGroup
	//
	// Count testflow numbers in group
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
	//       + name: groupId
	//         in: query
	//         description: group id  of test flow
	//         required: true
	//         type: string
	//
	//     Responses:
	//       200:
	//		 503: apiError
	group.GET("count/:groupId", func(c *gin.Context) {
		groupId, err := primitive.ObjectIDFromHex(c.Param("groupId"))
		if err != nil {
			c.Error(err)
			return
		}

		output, err := service.CountByGroup(ctx, groupId)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, output)
	})

	// swagger:route GET /testflow getTestFlow
	//
	// Get specific test case by condition.
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
	//       + name: name
	//         in: query
	//         description: name of test flow
	//         required: true
	//         type: string
	//       + name: id
	//         in: query
	//         description: id of test flow
	//         required: true
	//         type: string
	//
	//     Responses:
	//       200: testFlow
	//		 503: apiError
	group.GET("", func(c *gin.Context) {
		req := &models.GetTestFlowRequest{}
		err := c.ShouldBindQuery(req)
		if err != nil {
			c.Error(err)
			return
		}

		id, err := primitive.ObjectIDFromHex(req.ID)
		if err != nil {
			c.Error(err)
			return
		}

		output, err := service.Get(ctx, &repo.GetTestFlowParams{
			ID:   id,
			Name: req.Name,
		})
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, output)
	})

	// swagger:route POST /testflow saveTestFlow
	//
	// save test case, create if not exist
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
	//       + name: testflow
	//         in: body
	//         description: test flow json
	//         required: true
	//         type: testFlow
	//         allowEmpty:  false
	//
	//     Responses:
	//       200:
	//		 503: apiError
	group.POST("", func(c *gin.Context) {
		testFlow := models.TestFlow{}
		err := c.ShouldBindJSON(&testFlow)
		if err != nil {
			c.Error(err)
			return
		}

		id, err := service.Save(ctx, testFlow)
		if err != nil {
			c.Error(err)
			return
		}

		c.String(http.StatusOK, id.Hex())
	})

	// swagger:route DELETE /testflow/{id} deleteTestFlow
	//
	// Delete test flow by id
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
	//         description: id of test flow
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
		err = service.Delete(c, id)
		if err != nil {
			c.Error(err)
			return
		}

		c.Status(http.StatusOK)
	})

	// swagger:route POST /changegroup changetestflow
	//
	// change testflow group id
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
	//       + name: changGroupRequest
	//         in: body
	//         description: params with group id and testflows
	//         required: true
	//         type: changeTestflowGroupRequest
	//         allowEmpty:  false
	//
	//     Responses:
	//       200:
	//		 503: apiError
	group.POST("/changegroup", func(c *gin.Context) {
		changeTestflowGroup := models.ChangeTestflowGroupRequest{}
		err := c.ShouldBindJSON(&changeTestflowGroup)
		if err != nil {
			c.Error(err)
			return
		}

		err = service.ChangeTestflowGroup(ctx, repo.ChangeTestflowGroup(changeTestflowGroup))
		if err != nil {
			c.Error(err)
			return
		}

		c.Status(http.StatusOK)
	})
}

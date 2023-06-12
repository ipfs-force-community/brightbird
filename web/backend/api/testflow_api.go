package api

import (
	"context"
	"net/http"

	"github.com/hunjixin/brightbird/models"

	"github.com/gin-gonic/gin"
	"github.com/hunjixin/brightbird/repo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func RegisterTestFlowRouter(ctx context.Context, v1group *V1RouterGroup, service repo.ITestFlowRepo) {
	group := v1group.Group("/testflow")
	// swagger:route GET /testflow/list testflow listInGroupRequest
	//
	// Lists test flows.
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
	//       200: listTestFlowResp
	//		 503: apiError
	group.GET("list", func(c *gin.Context) {
		req := &models.ListInGroupRequest{}
		err := c.ShouldBindQuery(req)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		params := repo.ListTestFlowParams{
			Name: req.Name,
		}
		if req.GroupID != nil {
			params.GroupID, err = primitive.ObjectIDFromHex(*req.GroupID)
			if err != nil {
				c.Error(err) //nolint
				return
			}
		}

		output, err := service.List(ctx, models.PageReq[repo.ListTestFlowParams]{
			PageNum:  req.PageNum,
			PageSize: req.PageSize,
			Params:   params,
		})
		if err != nil {
			c.Error(err) //nolint
			return
		}

		c.JSON(http.StatusOK, output)
	})

	// swagger:route GET /testflow/count testflow countTestFlowRequest
	//
	// Count testflow numbers in group
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
	//       200:
	//		 503: apiError
	group.GET("count", func(c *gin.Context) {
		req := &models.CountTestFlowRequest{}
		err := c.ShouldBindQuery(req)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		params := &repo.CountTestFlowParams{
			Name: req.Name,
		}

		if req.GroupID != nil {
			groupID, err := primitive.ObjectIDFromHex(*req.GroupID)
			if err != nil {
				c.Error(err) //nolint
				return
			}
			params.GroupID = groupID
		}

		output, err := service.Count(ctx, params)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		c.JSON(http.StatusOK, output)
	})

	// swagger:route GET /testflow testflow getTestFlowRequest
	//
	// Get specific test case by condition.
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
	//       200: testFlow
	//		 503: apiError
	group.GET("", func(c *gin.Context) {
		req := &models.GetTestFlowRequest{}
		err := c.ShouldBindQuery(req)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		params := &repo.GetTestFlowParams{
			Name: req.Name,
		}
		if req.ID != nil {
			params.ID, err = primitive.ObjectIDFromHex(*req.ID)
			if err != nil {
				c.Error(err) //nolint
				return
			}
		}

		output, err := service.Get(ctx, params)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		c.JSON(http.StatusOK, output)
	})

	// swagger:route POST /testflow testflow saveTestFlow
	//
	// save test case, create if not exist
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
			c.Error(err) //nolint
			return
		}

		id, err := service.Save(ctx, testFlow)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		c.String(http.StatusOK, id.Hex())
	})

	// swagger:route DELETE /testflow/{id} testflow deleteTestFlow
	//
	// Delete test flow by id
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
			c.Error(err) //nolint
			return
		}
		err = service.Delete(c, id)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		c.Status(http.StatusOK)
	})

	// swagger:route POST /changegroup testflow changetestflow
	//
	// change testflow group id
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
			c.Error(err) //nolint
			return
		}

		err = service.ChangeTestflowGroup(ctx, changeTestflowGroup)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		c.Status(http.StatusOK)
	})
}

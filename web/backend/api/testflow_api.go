package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hunjixin/brightbird/repo"
	"github.com/hunjixin/brightbird/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BasePageReq struct {
	PageNum  int `form:"pageNum"`
	PageSize int `form:"pageSize"`
}

type BasePageResp struct {
	Total   int `json:"total"`
	Pages   int `json:"pages"`
	PageNum int `json:"pageNum"`
}

// ListInGroupRequest
// swagger:model listInGroupRequest
type ListInGroupRequest struct {
	BasePageReq
	// the group id of test flow
	// required: true
	GroupId string `form:"groupId" binding:"required"`
}

// GetTestFlowRequest
// swagger:model getTestFlowRequest
type GetTestFlowRequest = repo.GetTestFlowParams

// ListTestFlowResp
// swagger:model listTestFlowResp
type ListTestFlowResp struct {
	BasePageResp
	List []*types.TestFlow `json:"list"`
}

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
	group.GET("list", func(c *gin.Context) {
		req := &ListInGroupRequest{}
		err := c.ShouldBindQuery(req)
		if err != nil {
			c.Error(err)
			return
		}
		output, err := service.ListInGroup(ctx, &types.PageReq[string]{
			PageNum:  req.PageNum,
			PageSize: req.PageSize,
			Params:   req.GroupId,
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

	// swagger:route GET /testflow/{name} getTestFlow
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
	group.GET("", func(c *gin.Context) {
		req := &GetTestFlowRequest{}
		err := c.ShouldBindQuery(req)
		if err != nil {
			c.Error(err)
			return
		}

		output, err := service.Get(ctx, req)
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
	group.POST("", func(c *gin.Context) {
		testFlow := types.TestFlow{}
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
	//
	//     Responses:
	//       200:
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
}

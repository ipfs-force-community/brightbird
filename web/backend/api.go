package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
)

func RegisterCommonRouter(ctx context.Context, v1group *V1RouterGroup) {
	group := v1group.Group("/deploy")
	// swagger:route GET /version getVersion
	//
	// get backend version
	//
	//     Consumes:
	//
	//     Produces:
	//     - application/text
	//
	//     Schemes: http, https
	//
	//     Deprecated: false
	//
	//     Responses:
	group.GET("version", func(c *gin.Context) {
		c.JSON(http.StatusOK, version.Version())
	})
}

func RegisterGroupRouter(ctx context.Context, v1group *V1RouterGroup, service IGroupService) {
	group := v1group.Group("/group")

	// swagger:route GET /group listGroup
	//
	// Lists all group.
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
	//       200: []group
	group.GET("", func(c *gin.Context) {
		output, err := service.List(ctx)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, output)
	})

	// swagger:route GET /group/{name} getTestFlow
	//
	// Get specific group by name.
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
	//         in: path
	//         description: name of test flow
	//         required: true
	//         type: string
	//
	//     Responses:
	//       200: group
	group.GET(":name", func(c *gin.Context) {
		name := c.Param("name")
		output, err := service.Get(ctx, name)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, output)
	})

	// swagger:route POST /group saveCases
	//
	// Get specific test case by name.
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
	//       + name: group
	//         in: body
	//         description: group json
	//         required: true
	//         type: group
	//         allowEmpty:  false
	//
	//     Responses:
	//       200:
	group.POST("", func(c *gin.Context) {
		testFlow := types.Group{}
		err := c.ShouldBindJSON(&testFlow)
		if err != nil {
			c.Error(err)
			return
		}

		err = service.Save(ctx, testFlow)
		if err != nil {
			c.Error(err)
			return
		}

		c.Status(http.StatusOK)
	})
}

func RegisterDeployRouter(ctx context.Context, v1group *V1RouterGroup, service IPluginService) {
	group := v1group.Group("/deploy")

	// swagger:route GET /deploy/plugins listDeployPlugins
	//
	// Lists all deploy plugin.
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
	//       200: []pluginOut
	group.GET("plugins", func(c *gin.Context) {
		output, err := service.Plugins(c)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, output)
	})
}

func RegisterCasesRouter(ctx context.Context, v1group *V1RouterGroup, service ITestCaseService) {
	group := v1group.Group("/cases")

	// swagger:route GET /cases/plugins listCasesPlugins
	//
	// Lists all exec plugins.
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
	//       200: []pluginOut
	group.GET("plugins", func(c *gin.Context) {
		output, err := service.Plugins(c)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, output)
	})

	// swagger:route GET /cases/listall listAllTestFlows
	//
	// Lists all exec test flows.
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
	//       200: []testFlow
	group.GET("listall", func(c *gin.Context) {
		output, err := service.List(ctx)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, output)
	})

	// swagger:route GET /cases/list/ listTestFlowsInGroup
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
	//       200: listInGroupRequestResp
	group.GET("list", func(c *gin.Context) {
		req := &ListInGroupRequest{}
		err := c.ShouldBindQuery(req)
		if err != nil {
			c.Error(err)
			return
		}
		output, err := service.ListInGroup(ctx, req)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, output)
	})

	// swagger:route GET /cases/name/{name} getTestFlowByName
	//
	// Get specific test case by name.
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
	//         in: path
	//         description: name of test flow
	//         required: true
	//         type: string
	//
	//     Responses:
	//       200: testFlow
	group.GET("name/:name", func(c *gin.Context) {
		name := c.Param("name")
		output, err := service.GetByName(ctx, name)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, output)
	})

	// swagger:route GET /cases/id/{id} getTestFlowById
	//
	// Get specific test case by id.
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
	//       200: testFlow
	group.GET("id/:id", func(c *gin.Context) {
		id, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.Error(err)
			return
		}

		output, err := service.GetById(ctx, id)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, output)
	})

	// swagger:route POST /cases saveCases
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

		err = service.Save(ctx, testFlow)
		if err != nil {
			c.Error(err)
			return
		}

		c.Status(http.StatusOK)
	})
}

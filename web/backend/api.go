package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
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

	// swagger:route GET /cases/list listTestFlows
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
	group.GET("list", func(c *gin.Context) {
		output, err := service.List(ctx)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, output)
	})

	// swagger:route GET /cases/{name} getTestFlow
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
	group.GET(":name", func(c *gin.Context) {
		name := c.Param("name")
		output, err := service.Get(ctx, name)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, output)
	})

	// swagger:route POST /cases saveCases
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

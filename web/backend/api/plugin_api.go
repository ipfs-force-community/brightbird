package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hunjixin/brightbird/repo"
)

func RegisterDeployRouter(ctx context.Context, v1group *V1RouterGroup, service repo.IPluginService) {
	group := v1group.Group("/plugin")

	// swagger:route GET /plugin/deploy listDeployPlugins
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
	group.GET("deploy/list", func(c *gin.Context) {
		output, err := service.DeployPlugins(c)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, output)
	})

	// swagger:route GET /plugin/exec listExecPlugin
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
	group.GET("exec/list", func(c *gin.Context) {
		output, err := service.ExecPlugins(c)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, output)
	})

	// swagger:route GET /deploy/get/{name} getPluginByName
	//
	// Get deploy plugin by name.
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
	//         description: name of deploy
	//         required: true
	//         type: string
	//
	//     Responses:
	//       200: testFlow
	group.GET(":name", func(c *gin.Context) {
		name := c.Param("name")
		output, err := service.GetByName(c, name)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, output)
	})
}

package api

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/hunjixin/brightbird/web/backend/services"
	"net/http"
)

func RegisterDeployRouter(ctx context.Context, v1group *V1RouterGroup, service services.IPluginService) {
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

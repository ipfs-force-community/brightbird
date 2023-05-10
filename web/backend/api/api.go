package api

import (
	"context"
	"github.com/gin-gonic/gin"
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
	//       200: myString
	group.GET("version", func(c *gin.Context) {
		c.JSON(http.StatusOK, version.Version())
	})
}

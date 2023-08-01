package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ipfs-force-community/brightbird/version"
)

func RegisterCommonRouter(ctx context.Context, v1group *V1RouterGroup) {
	// swagger:route GET /version version getVersion
	//
	// get backend version
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
	//       200: myString
	//		 503: apiError
	v1group.GET("version", func(c *gin.Context) {
		c.JSON(http.StatusOK, version.Version())
	})
}

package runnerctl

import (
	"context"
	"errors"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ipfs-force-community/brightbird/types"
	logging "github.com/ipfs/go-log/v2"
	"go.uber.org/fx"
)

var log = logging.Logger("api")

type Params struct {
	fx.In
	Endpoint types.Endpoint
	Engin    *gin.Engine
	Shutdown types.Shutdown
}

type APIController struct {
	endpoint types.Endpoint
	engin    *gin.Engine
	shutdown types.Shutdown
}

func NewAPIController(params Params) *APIController {
	return &APIController{
		endpoint: params.Endpoint,
		engin:    params.Engin,
		shutdown: params.Shutdown,
	}
}

func SetupAPI(ctx context.Context, lc fx.Lifecycle, apiCtl *APIController) error {
	v1API := apiCtl.engin.Group("v1")
	v1API.POST("stop", func(e *gin.Context) {
		apiCtl.shutdown <- struct{}{}
		e.Writer.WriteHeader(http.StatusOK)
	})
	return listenAndWaitExit(ctx, lc, apiCtl)
}

func listenAndWaitExit(ctx context.Context, lc fx.Lifecycle, apiCtl *APIController) error {
	listener, err := net.Listen("tcp", string(apiCtl.endpoint))
	if err != nil {
		return err
	}

	log.Infof("Start listen api %s", listener.Addr().String())
	go func() {
		err = apiCtl.engin.RunListener(listener)
		if err != nil && errors.Is(err, http.ErrServerClosed) {
			log.Errorf("listen address fail %s", err)
		}
	}()

	lc.Append(fx.StartHook(func() {
		_ = listener.Close()
	}))

	return nil
}

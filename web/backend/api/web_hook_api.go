package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v51/github"
	"github.com/hunjixin/brightbird/web/backend/modules"
	logging "github.com/ipfs/go-log/v2"
)

var hookLogger = logging.Logger("hook_api")

func RegisterWebhookRouter(ctx context.Context, v1group *V1RouterGroup, webhookPubsub modules.WebHookPubsub) {
	webHookGroup := v1group.Group("webhook")

	webHookGroup.POST("", func(c *gin.Context) {
		event := &github.Event{}
		dec := json.NewDecoder(c.Request.Body)
		err := dec.Decode(event)
		if err != nil {
			c.Error(err)
			return
		}

		webhookPubsub.Pub(event, modules.WEB_HOOK_TOPIC)
		c.Status(http.StatusOK)
	})
}

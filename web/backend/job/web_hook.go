package job

import (
	"github.com/cskr/pubsub"
)

type WebhookHub struct {
	bus *pubsub.PubSub
}

func NewWebhookHub() *WebhookHub {
	return &WebhookHub{
		bus: pubsub.New(10),
	}
}

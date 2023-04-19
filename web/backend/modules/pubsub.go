package modules

import "github.com/cskr/pubsub"

const WEB_HOOK_TOPIC = "webhook"

type WebHookPubsub = *pubsub.PubSub

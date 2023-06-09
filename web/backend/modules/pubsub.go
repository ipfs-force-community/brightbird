package modules

import "github.com/cskr/pubsub"

const ReleaseTopic = "release"

const PRMergeTopic = "pr_merged"

type WebHookPubsub = *pubsub.PubSub

package modules

import "github.com/cskr/pubsub"

const RELEASE_TOPIC = "release"
const PR_MERGED_TOPIC = "pr_merged"

type WebHookPubsub = *pubsub.PubSub

package modules

import "github.com/cskr/pubsub"

const CREATE_TAG_TOPIC = "tag_create"
const PR_MERGED_TOPIC = "pr_merged"

type WebHookPubsub = *pubsub.PubSub

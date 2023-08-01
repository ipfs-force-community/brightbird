package repo

import (
	"github.com/ipfs-force-community/brightbird/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var sortModifyDesc = options.Find().SetAllowDiskUse(true).SetSort(bson.D{{Key: "modifiedtime", Value: -1}})

func PaginationAndSortByModifiyTimeDesc[T any](req models.PageReq[T]) *options.FindOptions {
	return options.Find().SetAllowDiskUse(true).SetSort(bson.D{{Key: "modifiedtime", Value: -1}}).SetSkip(req.Skip()).SetLimit(req.Take())
}

var sortNameDesc = options.Find().SetAllowDiskUse(true).SetSort(bson.D{{Key: "name", Value: -1}}) //nolint

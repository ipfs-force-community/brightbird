package repo

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/ipfs-force-community/brightbird/pkg/models"
)

var sortModifyDesc = options.Find().SetAllowDiskUse(true).SetSort(bson.D{{Key: "modifiedtime", Value: -1}})
var sortNameDesc = options.Find().SetAllowDiskUse(true).SetSort(bson.D{{Key: "name", Value: -1}}) //nolint
var sortNameAsc = options.Find().SetAllowDiskUse(true).SetSort(bson.D{{Key: "name", Value: 1}})   //nolint

func PaginationAndSortByModifiyTimeDesc[T any](req models.PageReq[T]) *options.FindOptions {
	return options.Find().SetAllowDiskUse(true).SetSort(bson.D{{Key: "modifiedtime", Value: -1}}).SetSkip(req.Skip()).SetLimit(req.Take())
}
func PaginationAndSortByNameAsc[T any](req models.PageReq[T]) *options.FindOptions {
	return options.Find().SetAllowDiskUse(true).SetSort(bson.D{{Key: "name", Value: 1}}).SetSkip(req.Skip()).SetLimit(req.Take())
}

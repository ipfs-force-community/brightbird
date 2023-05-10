package repo

import (
	"github.com/hunjixin/brightbird/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var sortModifyDesc = options.Find().SetAllowDiskUse(true).SetSort(bson.D{{"modifiedtime", -1}})

func PaginationAndSortByModifiyTimeDesc[T any](req models.PageReq[T]) *options.FindOptions {
	return options.Find().SetAllowDiskUse(true).SetSort(bson.D{{"modifiedtime", -1}}).SetSkip(int64(req.Skip())).SetLimit(int64(req.Take()))
}

var sortNameDesc = options.Find().SetAllowDiskUse(true).SetSort(bson.D{{"name", -1}})

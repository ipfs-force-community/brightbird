package repo

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var sortModifyDesc = options.Find().SetAllowDiskUse(true).SetSort(bson.D{{"basetime.modifiedtime", -1}})

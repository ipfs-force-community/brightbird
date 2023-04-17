package repo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

type ILogRepo interface {
	ListPodsInTest(context.Context, string) ([]string, error)
	GetPodLog(context.Context, string) ([]string, error)
}

type LogRepo struct {
	col *mongo.Collection
}

var _ ILogRepo = (*LogRepo)(nil)

func NewLogRepo(ctx context.Context, db *mongo.Database) (*LogRepo, error) {
	col := db.Collection("logs")
	_, err := col.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bsonx.Doc{{Key: "kubernetes.labels.testid", Value: bsonx.Int32(-1)}},
		},
		{
			Keys: bsonx.Doc{{Key: "kubernetes.pod_name", Value: bsonx.Int32(-1)}},
		},
		{
			Keys:    bsonx.Doc{{Key: "kubernetes.time", Value: bsonx.Int32(1)}},
			Options: options.Index().SetExpireAfterSeconds(60 * 60 * 24 * 7), //keep latest one week logs
		},
	})
	if err != nil {
		return nil, err
	}
	return &LogRepo{col: col}, nil
}

func (logRepo *LogRepo) ListPodsInTest(ctx context.Context, testid string) ([]string, error) {
	matchStage := bson.D{{"$match", bson.D{{"kubernetes.labels.testid", testid}}}}
	sortStage := bson.D{{"$sort", bson.D{{"time", 1}}}}
	groupStage := bson.D{
		{"$group",
			bson.D{
				{"_id", "$kubernetes.pod_name"},
				{"time", bson.M{"$first": "$time"}},
			},
		},
	}
	groupResultsCur, err := logRepo.col.Aggregate(ctx, mongo.Pipeline{matchStage, sortStage, groupStage, sortStage}, options.Aggregate().SetAllowDiskUse(true))
	if err != nil {
		return nil, err
	}

	var groupResults []bson.M
	if err = groupResultsCur.All(ctx, &groupResults); err != nil {
		return nil, err
	}
	podName := make([]string, len(groupResults))
	for index, r := range groupResults {
		podName[index] = r["_id"].(string)
	}
	return podName, nil
}

func (logRepo *LogRepo) GetPodLog(ctx context.Context, podName string) ([]string, error) {
	logResultCur, err := logRepo.col.Find(ctx, bson.M{"kubernetes.pod_name": podName}, options.Find().SetSort(bson.D{{"time", 1}}).SetProjection(bson.D{{"log", 1}, {"_id", 0}}).SetAllowDiskUse(true))
	if err != nil {
		return nil, err
	}

	var logResults []bson.M
	if err = logResultCur.All(ctx, &logResults); err != nil {
		return nil, err
	}

	logs := make([]string, len(logResults))
	for index, r := range logResults {
		logs[index] = r["log"].(string)
	}
	return logs, nil
}

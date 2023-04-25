package repo

import (
	"context"
	"time"

	"github.com/hunjixin/brightbird/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

type IJobRepo interface {
	List(context.Context) ([]*types.Job, error)
	Get(context.Context, primitive.ObjectID) (*types.Job, error)
	Save(context.Context, *types.Job) (primitive.ObjectID, error)
	Delete(ctx context.Context, id primitive.ObjectID) error
	IncExecCount(ctx context.Context, id primitive.ObjectID) (*types.Job, error)
}

var _ IJobRepo = (*JobRepo)(nil)

type JobRepo struct {
	jobCol *mongo.Collection
}

func NewJobRepo(ctx context.Context, db *mongo.Database) (*JobRepo, error) {
	col := db.Collection("jobs")
	_, err := col.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bsonx.Doc{{Key: "name", Value: bsonx.Int32(-1)}},
		},
	})
	if err != nil {
		return nil, err
	}
	return &JobRepo{jobCol: col}, nil
}

func (j *JobRepo) List(ctx context.Context) ([]*types.Job, error) {
	cur, err := j.jobCol.Find(ctx, bson.M{}, sortModifyDesc)
	if err != nil {
		return nil, err
	}

	jobs := []*types.Job{}
	err = cur.All(ctx, &jobs)
	if err != nil {
		return nil, err
	}
	return jobs, nil
}

func (j *JobRepo) Get(ctx context.Context, id primitive.ObjectID) (*types.Job, error) {
	tf := &types.Job{}
	err := j.jobCol.FindOne(ctx, bson.D{{"_id", id}}).Decode(tf)
	if err != nil {
		return nil, err
	}
	return tf, nil
}

func (j *JobRepo) IncExecCount(ctx context.Context, id primitive.ObjectID) (*types.Job, error) {
	tf := &types.Job{}

	inc := bson.M{
		"$inc": bson.M{
			"execcount": 1,
		},
	}
	err := j.jobCol.FindOneAndUpdate(ctx, bson.D{{"_id", id}}, inc).Decode(tf)
	if err != nil {
		return nil, err
	}
	return tf, nil
}

func (j *JobRepo) Save(ctx context.Context, job *types.Job) (primitive.ObjectID, error) {
	if job.ID.IsZero() {
		job.ID = primitive.NewObjectID()
	}

	count, err := j.jobCol.CountDocuments(ctx, bson.D{{"_id", job.ID}})
	if err != nil {
		return primitive.ObjectID{}, err
	}
	if count == 0 {
		job.BaseTime.CreateTime = time.Now().Unix()
		job.BaseTime.ModifiedTime = time.Now().Unix()
	} else {
		job.BaseTime.ModifiedTime = time.Now().Unix()
	}

	update := bson.M{
		"$set": job,
	}
	_, err = j.jobCol.UpdateOne(ctx, bson.D{{"name", job.Name}}, update, options.Update().SetUpsert(true))
	if err != nil {
		return primitive.ObjectID{}, err
	}
	return job.ID, nil
}

func (j *JobRepo) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := j.jobCol.DeleteOne(ctx, bson.D{{"_id", id}})
	if err != nil {
		return err
	}
	return nil
}

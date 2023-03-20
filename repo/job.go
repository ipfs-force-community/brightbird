package repo

import (
	"context"
	"github.com/hunjixin/brightbird/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IJobRepo interface {
	List(context.Context) ([]*types.Job, error)
	Get(context.Context, primitive.ObjectID) (*types.Job, error)
	Save(context.Context, types.Job) error
	Delete(ctx context.Context, id primitive.ObjectID) error
}

var _ IJobRepo = (*JobRepo)(nil)

type JobRepo struct {
	jobCol *mongo.Collection
}

func NewJobRepo(db *mongo.Database) *JobRepo {
	return &JobRepo{jobCol: db.Collection("jobs")}
}

func (j *JobRepo) List(ctx context.Context) ([]*types.Job, error) {
	cur, err := j.jobCol.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	var tf []*types.Job
	err = cur.All(ctx, &tf)
	if err != nil {
		return nil, err
	}
	return tf, nil
}

func (j *JobRepo) Get(ctx context.Context, id primitive.ObjectID) (*types.Job, error) {
	tf := &types.Job{}
	err := j.jobCol.FindOne(ctx, bson.D{{"_id", id}}).Decode(tf)
	if err != nil {
		return nil, err
	}
	return tf, nil
}

func (j *JobRepo) Save(ctx context.Context, job types.Job) error {
	if job.ID.IsZero() {
		job.ID = primitive.NewObjectID()
	}
	update := bson.M{
		"$set": job,
	}
	_, err := j.jobCol.UpdateOne(ctx, bson.D{{"Name", job.Name}}, update, options.Update().SetUpsert(true))
	if err != nil {
		return err
	}
	return nil
}

func (j *JobRepo) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := j.jobCol.DeleteOne(ctx, bson.D{{"_id", id}})
	if err != nil {
		return err
	}
	return nil
}

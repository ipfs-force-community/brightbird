package repo

import (
	"context"
	"time"

	"github.com/hunjixin/brightbird/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

type CountJobParams struct {
	ID   primitive.ObjectID
	Name *string
}

type IJobRepo interface {
	Count(context.Context, *CountJobParams) (int64, error)
	List(context.Context) ([]*models.Job, error)
	Get(context.Context, primitive.ObjectID) (*models.Job, error)
	Save(context.Context, *models.Job) (primitive.ObjectID, error)
	Delete(ctx context.Context, id primitive.ObjectID) error
	IncExecCount(ctx context.Context, id primitive.ObjectID) (*models.Job, error)
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

func (j *JobRepo) List(ctx context.Context) ([]*models.Job, error) {
	cur, err := j.jobCol.Find(ctx, bson.M{}, sortModifyDesc)
	if err != nil {
		return nil, err
	}

	jobs := []*models.Job{}
	err = cur.All(ctx, &jobs)
	if err != nil {
		return nil, err
	}
	return jobs, nil
}

func (j *JobRepo) Count(ctx context.Context, params *CountJobParams) (int64, error) {
	filter := bson.M{}
	if params.Name != nil {
		filter["name"] = params.Name
	}
	if !params.ID.IsZero() {
		filter["_id"] = params.ID
	}
	return j.jobCol.CountDocuments(ctx, filter)
}

func (j *JobRepo) Get(ctx context.Context, id primitive.ObjectID) (*models.Job, error) {
	tf := &models.Job{}
	err := j.jobCol.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(tf)
	if err != nil {
		return nil, err
	}
	return tf, nil
}

func (j *JobRepo) IncExecCount(ctx context.Context, id primitive.ObjectID) (*models.Job, error) {
	tf := &models.Job{}

	inc := bson.M{
		"$inc": bson.M{
			"execcount": 1,
		},
	}
	err := j.jobCol.FindOneAndUpdate(ctx, bson.D{{Key: "_id", Value: id}}, inc).Decode(tf)
	if err != nil {
		return nil, err
	}
	return tf, nil
}

func (j *JobRepo) Save(ctx context.Context, job *models.Job) (primitive.ObjectID, error) {
	if job.ID.IsZero() {
		job.ID = primitive.NewObjectID()
	}

	count, err := j.jobCol.CountDocuments(ctx, bson.D{{Key: "_id", Value: job.ID}})
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
	_, err = j.jobCol.UpdateOne(ctx, bson.D{{Key: "name", Value: job.Name}}, update, options.Update().SetUpsert(true))
	if err != nil {
		return primitive.ObjectID{}, err
	}
	return job.ID, nil
}

func (j *JobRepo) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := j.jobCol.DeleteOne(ctx, bson.D{{Key: "_id", Value: id}})
	if err != nil {
		return err
	}
	return nil
}

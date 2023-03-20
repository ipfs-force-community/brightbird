package repo

import (
	"context"
	"github.com/hunjixin/brightbird/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ITaskRepo interface {
	List(context.Context) ([]*types.Task, error)
	ListInJob(context.Context, primitive.ObjectID) ([]*types.Task, error)
	UpdateVersion(ctx context.Context, id primitive.ObjectID, versionMap map[string]string) error
	Get(context.Context, primitive.ObjectID) (*types.Task, error)
	Save(context.Context, types.Task) error
	Delete(ctx context.Context, id primitive.ObjectID) error
}

var _ ITaskRepo = (*TaskRepo)(nil)

func NewTaskRepo(db *mongo.Database) *TaskRepo {
	return &TaskRepo{taskCol: db.Collection("tasks")}
}

type TaskRepo struct {
	taskCol *mongo.Collection
}

func (j *TaskRepo) ListInJob(ctx context.Context, jobId primitive.ObjectID) ([]*types.Task, error) {
	cur, err := j.taskCol.Find(ctx, bson.D{{"jobId", jobId}})
	if err != nil {
		return nil, err
	}

	var tasks []*types.Task
	err = cur.All(ctx, tasks)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (j *TaskRepo) List(ctx context.Context) ([]*types.Task, error) {
	cur, err := j.taskCol.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	var tf []*types.Task
	err = cur.All(ctx, &tf)
	if err != nil {
		return nil, err
	}
	return tf, nil
}

func (j *TaskRepo) Get(ctx context.Context, id primitive.ObjectID) (*types.Task, error) {
	tf := &types.Task{}
	err := j.taskCol.FindOne(ctx, bson.D{{"_id", id}}).Decode(tf)
	if err != nil {
		return nil, err
	}
	return tf, nil
}

func (j *TaskRepo) Save(ctx context.Context, task types.Task) error {
	if task.ID.IsZero() {
		task.ID = primitive.NewObjectID()
	}
	_, err := j.taskCol.InsertOne(ctx, task)
	if err != nil {
		return err
	}
	return nil
}

func (j *TaskRepo) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := j.taskCol.DeleteOne(ctx, bson.D{{"_id", id}})
	if err != nil {
		return err
	}
	return nil
}

func (j *TaskRepo) UpdateVersion(ctx context.Context, id primitive.ObjectID, versionMap map[string]string) error {
	update := bson.M{
		"$set": bson.M{
			"versions": versionMap,
		},
	}
	_, err := j.taskCol.UpdateByID(ctx, id, update)
	if err != nil {
		return err
	}
	return nil
}

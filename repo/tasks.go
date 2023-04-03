package repo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/hunjixin/brightbird/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ListParams struct {
	JobId primitive.ObjectID `json:"jobId"`
	State []types.State      `json:"state"`
}

type ITaskRepo interface {
	List(context.Context, ListParams) ([]*types.Task, error)
	UpdateVersion(ctx context.Context, id primitive.ObjectID, versionMap map[string]string) error
	MarkState(ctx context.Context, id primitive.ObjectID, state types.State, msg ...string) error
	UpdatePodRunning(ctx context.Context, id primitive.ObjectID, name string) error
	Get(context.Context, primitive.ObjectID) (*types.Task, error)
	Save(context.Context, *types.Task) (primitive.ObjectID, error)
	Delete(ctx context.Context, id primitive.ObjectID) error
}

var _ ITaskRepo = (*TaskRepo)(nil)

func NewTaskRepo(db *mongo.Database) *TaskRepo {
	return &TaskRepo{taskCol: db.Collection("tasks")}
}

type TaskRepo struct {
	taskCol *mongo.Collection
}

func (j *TaskRepo) List(ctx context.Context, params ListParams) ([]*types.Task, error) {
	filter := bson.D{}
	if !params.JobId.IsZero() {
		filter = append(filter, bson.E{Key: "jobid", Value: params.JobId})
	}

	if len(params.State) > 0 {
		filter = append(filter, bson.E{Key: "state", Value: bson.M{"$in": params.State}})
	}

	cur, err := j.taskCol.Find(ctx, filter)
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

func (j *TaskRepo) MarkState(ctx context.Context, id primitive.ObjectID, state types.State, logs ...string) error {
	update := bson.M{
		"$set": bson.M{
			"state": state,
		},
		"$push": bson.M{
			"logs": bson.M{
				"$each": logs,
			},
		},
	}

	_, err := j.taskCol.UpdateByID(ctx, id, update)
	return err
}

func (j *TaskRepo) UpdatePodRunning(ctx context.Context, id primitive.ObjectID, podName string) error {
	update := bson.M{
		"$set": bson.M{
			"state":   types.Running,
			"podname": podName,
		},
		"$push": bson.M{
			"logs": "submit testrunner successfully",
		},
	}

	_, err := j.taskCol.UpdateByID(ctx, id, update)
	return err
}

func (j *TaskRepo) Save(ctx context.Context, task *types.Task) (primitive.ObjectID, error) {
	if task.ID.IsZero() {
		task.ID = primitive.NewObjectID()
	}

	if task.Logs == nil {
		task.Logs = []string{} // init logs as aray
	}
	count, err := j.taskCol.CountDocuments(ctx, bson.D{{"_id", task.ID}})
	if err != nil {
		return primitive.ObjectID{}, err
	}
	if count == 0 {
		task.BaseTime.CreateTime = time.Now().Unix()
		task.BaseTime.ModifiedTime = time.Now().Unix()
	} else {
		task.BaseTime.ModifiedTime = time.Now().Unix()
	}

	update := bson.M{
		"$set": task,
	}
	_, err = j.taskCol.UpdateOne(ctx, bson.D{{"_id", task.ID}}, update, options.Update().SetUpsert(true))
	if err != nil {
		return primitive.ObjectID{}, err
	}

	return task.ID, nil
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

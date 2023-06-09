package repo

import (
	"context"
	"time"

	"github.com/hunjixin/brightbird/models"

	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ListTaskParams struct {
	JobID      primitive.ObjectID `form:"jobId"`
	State      []models.State     `form:"state"`
	CreateTime *int64             `form:"createtime"`
}

type ITaskRepo interface {
	List(context.Context, models.PageReq[ListTaskParams]) (*models.PageResp[*models.Task], error)
	UpdateCommitMap(ctx context.Context, id primitive.ObjectID, versionMap map[string]string) error
	MarkState(ctx context.Context, id primitive.ObjectID, state models.State, msg ...string) error
	UpdatePodRunning(ctx context.Context, id primitive.ObjectID, name string) error
	Get(context.Context, primitive.ObjectID) (*models.Task, error)
	Save(context.Context, *models.Task) (primitive.ObjectID, error)
	Delete(ctx context.Context, id primitive.ObjectID) error
}

var _ ITaskRepo = (*TaskRepo)(nil)

func NewTaskRepo(ctx context.Context, db *mongo.Database) (*TaskRepo, error) {
	col := db.Collection("tasks")
	_, err := col.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bsonx.Doc{{Key: "jobid", Value: bsonx.Int32(-1)}},
		},
		{
			Keys: bsonx.Doc{{Key: "state", Value: bsonx.Int32(-1)}},
		},
		{
			Keys: bsonx.Doc{
				{Key: "jobid", Value: bsonx.Int32(-1)},
				{Key: "state", Value: bsonx.Int32(-1)},
			},
		},
		{
			Keys: bsonx.Doc{
				{Key: "jobid", Value: bsonx.Int32(-1)},
				{Key: "state", Value: bsonx.Int32(-1)},
				{Key: "createtime", Value: bsonx.Int32(-1)},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	return &TaskRepo{taskCol: col}, nil
}

type TaskRepo struct {
	taskCol *mongo.Collection
}

func (j *TaskRepo) List(ctx context.Context, params models.PageReq[ListTaskParams]) (*models.PageResp[*models.Task], error) {
	filter := bson.D{}
	if !params.Params.JobID.IsZero() {
		filter = append(filter, bson.E{Key: "jobid", Value: params.Params.JobID})
	}

	if len(params.Params.State) > 0 {
		filter = append(filter, bson.E{Key: "state", Value: bson.M{"$in": params.Params.State}})
	}

	if params.Params.CreateTime != nil {
		filter = append(filter, bson.E{Key: "createtime", Value: bson.M{"$gt": *params.Params.CreateTime}})
	}

	count, err := j.taskCol.CountDocuments(ctx, filter)
	if err != nil {
		return nil, err
	}

	cur, err := j.taskCol.Find(ctx, filter, PaginationAndSortByModifiyTimeDesc(params))
	if err != nil {
		return nil, err
	}

	tasks := []*models.Task{} //ensure lisit have value convient for front pages
	err = cur.All(ctx, &tasks)
	if err != nil {
		return nil, err
	}

	return &models.PageResp[*models.Task]{
		Total:   count,
		Pages:   (count + params.PageSize - 1) / params.PageSize,
		PageNum: params.PageNum,
		List:    tasks,
	}, nil
}

func (j *TaskRepo) Get(ctx context.Context, id primitive.ObjectID) (*models.Task, error) {
	tf := &models.Task{}
	err := j.taskCol.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(tf)
	if err != nil {
		return nil, err
	}
	return tf, nil
}

func (j *TaskRepo) MarkState(ctx context.Context, id primitive.ObjectID, state models.State, logs ...string) error {
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
			"state":   models.Running,
			"podname": podName,
		},
		"$push": bson.M{
			"logs": "submit testrunner successfully",
		},
	}

	_, err := j.taskCol.UpdateByID(ctx, id, update)
	return err
}

func (j *TaskRepo) Save(ctx context.Context, task *models.Task) (primitive.ObjectID, error) {
	if task.ID.IsZero() {
		task.ID = primitive.NewObjectID()
	}

	if task.Logs == nil {
		task.Logs = []string{} // init logs as aray
	}
	count, err := j.taskCol.CountDocuments(ctx, bson.D{{Key: "_id", Value: task.ID}})
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
	_, err = j.taskCol.UpdateOne(ctx, bson.D{{Key: "_id", Value: task.ID}}, update, options.Update().SetUpsert(true))
	if err != nil {
		return primitive.ObjectID{}, err
	}

	return task.ID, nil
}

func (j *TaskRepo) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := j.taskCol.DeleteOne(ctx, bson.D{{Key: "_id", Value: id}})
	if err != nil {
		return err
	}
	return nil
}

func (j *TaskRepo) DeleteByJobId(ctx context.Context, jobId primitive.ObjectID) error {
	_, err := j.taskCol.DeleteMany(ctx, bson.D{{Key: "jobid", Value: jobId}})
	if err != nil {
		return err
	}
	return nil
}

func (j *TaskRepo) UpdateCommitMap(ctx context.Context, id primitive.ObjectID, versionMap map[string]string) error {
	update := bson.M{
		"$set": bson.M{
			"commitmap": versionMap,
		},
	}
	_, err := j.taskCol.UpdateByID(ctx, id, update)
	if err != nil {
		return err
	}
	return nil
}

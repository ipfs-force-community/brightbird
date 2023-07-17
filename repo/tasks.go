package repo

import (
	"context"
	"fmt"
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
	CountAllAmount(ctx context.Context, state ...models.State) (int64, error)
	TaskAmountOfJobLast2Week(ctx context.Context) (map[string][]int, []string, error)
	JobPassRateTop3Today(ctx context.Context) ([]string, []string, error)
	JobFailureRatiobLast2Week(ctx context.Context) (map[string]int, error)
	TasktPassRateLast30Days(ctx context.Context) ([]string, []int32, error)
	JobPassRateLast30Days(ctx context.Context) (map[string][]int, []string, error)
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

func (j *TaskRepo) CountAllAmount(ctx context.Context, state ...models.State) (int64, error) {
	filter := bson.M{}
	if len(state) > 0 {
		filter["state"] = state[0]
	}

	count, err := j.taskCol.CountDocuments(ctx, filter)
	if err != nil {
		return -1, err
	}
	return count, nil
}

func (j *TaskRepo) TaskAmountOfJobLast2Week(ctx context.Context) (map[string][]int, []string, error) {
	endD := time.Now().Unix()
	startD := time.Now().AddDate(0, 0, -14).Unix()

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"createtime": bson.M{
					"$gt": startD,
					"$lt": endD,
				},
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"jobid": "$jobid",
					"date":  "$createtime",
				},
				"task_count": bson.M{"$sum": 1},
			},
		},
		{
			"$group": bson.M{
				"_id":        "$_id.jobid",
				"date":       bson.M{"$addToSet": "$_id.date"},
				"task_count": bson.M{"$addToSet": "$task_count"},
			},
		},
		{
			"$project": bson.M{
				"_id":         0,
				"jobid":       "$_id",
				"dates":       "$date",
				"task_counts": bson.M{"$first": "$task_count"},
			},
		},
	}

	cursor, err := j.taskCol.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, nil, err
	}
	defer cursor.Close(ctx)

	now := time.Now().UTC()
	startDate := now.AddDate(0, 0, -14)
	dateArray := make([]string, 0)
	for date := startDate; date.Before(now); date = date.AddDate(0, 0, 1) {
		dateStr := date.Format("01-02")
		dateArray = append(dateArray, dateStr)
	}

	jobIDHashTable := make(map[string][]int)

	for cursor.Next(ctx) {
		var result bson.M
		if err := cursor.Decode(&result); err != nil {
			return nil, nil, err
		}

		jobID := result["jobid"].(primitive.ObjectID).Hex()

		dates := result["dates"].(primitive.A)
		datesSlice := []interface{}(dates)
		timestamp := datesSlice[0].(int64)
		date := time.Unix(timestamp, 0)
		diff := date.Sub(startDate).Hours() / 24
		daysDiff := int(diff)

		taskCount := result["task_counts"].(int32)

		taskCountSlice := make([]int, 14)
		for i := range taskCountSlice {
			taskCountSlice[i] = 0
		}
		taskCountSlice[daysDiff] = int(taskCount)

		jobIDHashTable[jobID] = taskCountSlice

	}

	if err := cursor.Err(); err != nil {
		return nil, nil, err
	}

	for jobID, counts := range jobIDHashTable {
		fmt.Printf("JobID: %s, Counts: %v\n", jobID, counts)
	}

	return jobIDHashTable, dateArray, nil
}

func (j *TaskRepo) JobPassRateTop3Today(ctx context.Context) ([]string, []string, error) {
	endD := time.Now().Unix()
	startD := time.Now().AddDate(0, 0, -1).Unix()

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"createtime": bson.M{
					"$gt": startD,
					"$lt": endD,
				},
			},
		},
		{
			"$group": bson.M{
				"_id": "$jobid",
				"pass_count": bson.M{
					"$sum": bson.M{
						"$cond": bson.M{
							"if":   bson.M{"$eq": []interface{}{"$state", 5}},
							"then": 1,
							"else": 0,
						},
					},
				},
				"total_count": bson.M{"$sum": 1},
			},
		},
		{
			"$project": bson.M{
				"jobid": "$_id",
				"pass_rate": bson.M{"$concat": []interface{}{
					bson.M{"$toString": bson.M{"$multiply": []interface{}{
						bson.M{"$divide": []interface{}{
							"$pass_count",
							bson.M{"$cond": []interface{}{
								bson.M{"$ne": []interface{}{"$total_count", 0}},
								"$total_count",
								1,
							}},
						}},
						100,
					}}},
					"%",
				}},
			},
		},
		{
			"$sort": bson.M{"pass_rate": -1},
		},
	}

	cursor, err := j.taskCol.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, nil, err
	}
	defer cursor.Close(ctx)

	var jobNames []string
	var passRates []string
	for cursor.Next(ctx) {
		var result bson.M
		if err := cursor.Decode(&result); err != nil {
			return nil, nil, err
		}

		jobName := result["jobid"].(primitive.ObjectID).Hex()
		passRate := result["pass_rate"].(string)

		jobNames = append(jobNames, jobName)
		passRates = append(passRates, passRate)
	}

	if err := cursor.Err(); err != nil {
		return nil, nil, err
	}

	return jobNames, passRates, nil
}

func (j *TaskRepo) JobFailureRatiobLast2Week(ctx context.Context) (map[string]int, error) {
	endD := time.Now().Unix()
	startD := time.Now().AddDate(0, 0, -14).Unix()

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"state": 4,
				"createtime": bson.M{
					"$gt": startD,
					"$lt": endD,
				},
			},
		},
		{
			"$group": bson.M{
				"_id":           "$jobid",
				"failure_count": bson.M{"$sum": 1},
			},
		},
	}

	cursor, err := j.taskCol.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	jobIDHashTable := make(map[string]int)
	for _, result := range results {
		jobID := result["_id"].(primitive.ObjectID).Hex()
		failureCount := result["failure_count"].(int32)

		jobIDHashTable[jobID] = int(failureCount)
	}

	return jobIDHashTable, nil
}

func (j *TaskRepo) TasktPassRateLast30Days(ctx context.Context) ([]string, []int32, error) {
	passTaskArray := make([]int32, 30)

	endD := time.Now().Unix()
	startD := time.Now().AddDate(0, 0, -30).Unix()

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"createtime": bson.M{
					"$gte": startD,
					"$lt":  endD,
				},
				"state": 5,
			},
		},
		{
			"$group": bson.M{
				"_id":        "$createtime",
				"pass_count": bson.M{"$sum": 1},
			},
		},
		{
			"$sort": bson.M{"_id": 1},
		},
	}

	cursor, err := j.taskCol.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, nil, err
	}
	defer cursor.Close(ctx)

	dateArray := make([]string, 0)
	now := time.Now().UTC()
	startDate := now.AddDate(0, 0, -30)
	for date := startDate; date.Before(now); date = date.AddDate(0, 0, 1) {
		dateStr := date.Format("01-02")
		dateArray = append(dateArray, dateStr)
	}

	for cursor.Next(ctx) {
		var result bson.M
		if err := cursor.Decode(&result); err != nil {
			return nil, nil, err
		}

		timestamp := result["_id"].(int64)
		t := time.Unix(timestamp, 0)
		diff := t.Sub(startDate).Hours() / 24
		daysDiff := int(diff)

		count := result["pass_count"].(int32)

		if err != nil {
			return nil, nil, err
		}

		passTaskArray[daysDiff] += count
	}

	if err := cursor.Err(); err != nil {
		return nil, nil, err
	}

	return dateArray, passTaskArray, nil
}

func (j *TaskRepo) JobPassRateLast30Days(ctx context.Context) (map[string][]int, []string, error) {
	endD := time.Now().Unix()
	startD := time.Now().AddDate(0, 0, -30).Unix()

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"createtime": bson.M{
					"$gte": startD,
					"$lt":  endD,
				},
				"state": 5,
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"jobid": "$jobid",
					"date":  "$createtime",
				},
				"task_count": bson.M{"$sum": 1},
			},
		},
		{
			"$group": bson.M{
				"_id":        "$_id.jobid",
				"date":       bson.M{"$addToSet": "$_id.date"},
				"task_count": bson.M{"$addToSet": "$task_count"},
			},
		},
		{
			"$project": bson.M{
				"_id":         0,
				"jobid":       "$_id",
				"dates":       "$date",
				"task_counts": bson.M{"$first": "$task_count"},
			},
		},
	}

	cursor, err := j.taskCol.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, nil, err
	}
	defer cursor.Close(ctx)

	now := time.Now().UTC()
	startDate := now.AddDate(0, 0, -30)
	dateArray := make([]string, 0)
	for date := startDate; date.Before(now); date = date.AddDate(0, 0, 1) {
		dateStr := date.Format("01-02")
		dateArray = append(dateArray, dateStr)
	}

	jobIDHashTable := make(map[string][]int)

	for cursor.Next(ctx) {
		var result bson.M
		if err := cursor.Decode(&result); err != nil {
			return nil, nil, err
		}

		jobID := result["jobid"].(primitive.ObjectID).Hex()

		dates := result["dates"].(primitive.A)
		datesSlice := []interface{}(dates)
		timestamp := datesSlice[0].(int64)
		date := time.Unix(timestamp, 0)
		diff := date.Sub(startDate).Hours() / 24
		daysDiff := int(diff)

		taskCount := result["task_counts"].(int32)

		taskCountSlice := make([]int, 30)
		for i := range taskCountSlice {
			taskCountSlice[i] = 0
		}
		taskCountSlice[daysDiff] = int(taskCount)

		jobIDHashTable[jobID] = taskCountSlice

	}

	if err := cursor.Err(); err != nil {
		return nil, nil, err
	}

	for jobID, counts := range jobIDHashTable {
		fmt.Printf("JobID: %s, Counts: %v\n", jobID, counts)
	}

	return jobIDHashTable, dateArray, nil
}

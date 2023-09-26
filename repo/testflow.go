package repo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ipfs-force-community/brightbird/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type GetTestFlowParams struct {
	ID   primitive.ObjectID `form:"id"`
	Name *string            `form:"name"`
}

type ListTestFlowParams struct {
	GroupID primitive.ObjectID
	Name    *string
}

type CountTestFlowParams struct {
	GroupID primitive.ObjectID
	Name    *string
}

type ChangeTestflowGroup = models.ChangeTestflowGroupRequest

type ITestFlowRepo interface {
	Get(context.Context, *GetTestFlowParams) (*models.TestFlow, error)
	List(ctx context.Context, req models.PageReq[ListTestFlowParams]) (*models.PageResp[models.TestFlow], error)
	Save(context.Context, models.TestFlow) (primitive.ObjectID, error)
	Count(ctx context.Context, params *CountTestFlowParams) (int64, error)
	Copy(ctx context.Context, id primitive.ObjectID, name string, GroupId primitive.ObjectID) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	ChangeTestflowGroup(ctx context.Context, params ChangeTestflowGroup) error
}

type TestFlowRepo struct {
	caseCol *mongo.Collection
}

func NewTestFlowRepo(ctx context.Context, db *mongo.Database) (*TestFlowRepo, error) {
	col := db.Collection("testflows")
	_, err := col.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "name", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "groupid", Value: -1}},
		},
		{
			Keys: bson.D{
				{Key: "name", Value: -1},
				{Key: "groupid", Value: -1},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	return &TestFlowRepo{caseCol: col}, nil
}

func (c *TestFlowRepo) List(ctx context.Context, params models.PageReq[ListTestFlowParams]) (*models.PageResp[models.TestFlow], error) {
	filter := bson.D{}
	if !params.Params.GroupID.IsZero() {
		filter = append(filter, bson.E{Key: "groupid", Value: params.Params.GroupID})
	}
	if params.Params.Name != nil {
		filter = append(filter, bson.E{Key: "name", Value: bson.M{"$regex": primitive.Regex{
			Pattern: *params.Params.Name,
			Options: "im",
		}}})
	}
	count, err := c.caseCol.CountDocuments(ctx, filter)
	if err != nil {
		return nil, err
	}

	cur, err := c.caseCol.Find(ctx, filter, PaginationAndSortByModifiyTimeDesc(params))
	if err != nil {
		return nil, err
	}

	tf := []models.TestFlow{}
	err = cur.All(ctx, &tf)
	if err != nil {
		return nil, err
	}

	return &models.PageResp[models.TestFlow]{
		Total:   count,
		Pages:   (count + params.PageSize - 1) / params.PageSize,
		PageNum: params.PageNum,
		List:    tf,
	}, nil
}

func (c *TestFlowRepo) Get(ctx context.Context, params *GetTestFlowParams) (*models.TestFlow, error) {
	filter := bson.M{}
	if params.Name != nil {
		filter["name"] = *params.Name
	}
	if !params.ID.IsZero() {
		filter["_id"] = params.ID
	}

	tf := &models.TestFlow{}
	err := c.caseCol.FindOne(ctx, filter).Decode(tf)
	if err != nil {
		return nil, err
	}
	return tf, nil
}

func (c *TestFlowRepo) Count(ctx context.Context, params *CountTestFlowParams) (int64, error) {
	filter := bson.M{}
	if params.Name != nil {
		filter["name"] = params.Name
	}
	if !params.GroupID.IsZero() {
		filter["groupid"] = params.GroupID
	}
	return c.caseCol.CountDocuments(ctx, filter)
}

func (c *TestFlowRepo) Save(ctx context.Context, tf models.TestFlow) (primitive.ObjectID, error) {
	if tf.ID.IsZero() {
		tf.ID = primitive.NewObjectID()
	}

	count, err := c.caseCol.CountDocuments(ctx, bson.D{{Key: "_id", Value: tf.ID}})
	if err != nil {
		return primitive.ObjectID{}, err
	}
	if count == 0 {
		tf.BaseTime.CreateTime = time.Now().Unix()
		tf.BaseTime.ModifiedTime = time.Now().Unix()
	} else {
		tf.BaseTime.ModifiedTime = time.Now().Unix()
	}

	update := bson.M{
		"$set": tf,
	}
	_, err = c.caseCol.UpdateOne(ctx, bson.D{{Key: "name", Value: tf.Name}}, update, options.Update().SetUpsert(true))
	if err != nil {
		return primitive.ObjectID{}, err
	}

	return tf.ID, nil
}

func (c *TestFlowRepo) Delete(ctx context.Context, id primitive.ObjectID) error {
	tf := &models.TestFlow{}
	err := c.caseCol.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(tf)
	if err != nil || err == mongo.ErrNoDocuments {
		return fmt.Errorf("test flow (%d) not exis", id)
	}

	_, err = c.caseCol.DeleteOne(ctx, bson.D{{Key: "_id", Value: id}})
	if err != nil {
		return err
	}
	return nil
}

func (c *TestFlowRepo) Copy(ctx context.Context, id primitive.ObjectID, name string, GroupId primitive.ObjectID) error {
	tf := &models.TestFlow{}
	err := c.caseCol.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(tf)
	if err != nil {
		return err
	}

	if tf.Name == name {
		return fmt.Errorf("name %s was already exit", name)
	}

	tf.Graph = strings.Replace(tf.Graph, fmt.Sprintf("name: %s\n", tf.Name), fmt.Sprintf("name: %s\n", name), 1) //todo https://github.com/go-yaml/yaml/issues/698
	tf.ID = primitive.NewObjectID()
	tf.Name = name
	tf.CreateTime = time.Now().Unix()
	tf.ModifiedTime = time.Now().Unix()
	tf.GroupId = GroupId

	_, err = c.caseCol.InsertOne(ctx, tf)
	return err
}

func (c *TestFlowRepo) ChangeTestflowGroup(ctx context.Context, params ChangeTestflowGroup) error {
	var updateModels []mongo.WriteModel
	for _, testflowID := range params.TestflowIDs {
		update := bson.M{
			"$set": bson.D{{Key: "groupid", Value: params.GroupID}},
		}
		updateModels = append(updateModels, mongo.NewUpdateOneModel().SetFilter(bson.D{{Key: "_id", Value: testflowID}}).SetUpdate(update))
	}

	_, err := c.caseCol.BulkWrite(ctx, updateModels)
	return err
}

package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/hunjixin/brightbird/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type GetTestFlowParams struct {
	ID   primitive.ObjectID `form:"id"`
	Name string             `form:"name"`
}

type ListTestFlowParams struct {
	GroupID primitive.ObjectID
}

type ITestFlowRepo interface {
	Get(context.Context, *GetTestFlowParams) (*types.TestFlow, error)
	List(ctx context.Context, req *types.PageReq[ListTestFlowParams]) (*types.PageResp[types.TestFlow], error)
	Save(context.Context, types.TestFlow) (primitive.ObjectID, error)
	CountByGroup(ctx context.Context, groupId primitive.ObjectID) (int64, error)
	Delete(ctx context.Context, id primitive.ObjectID) error
}

type TestFlowRepo struct {
	caseCol *mongo.Collection
}

func NewTestFlowRepo(db *mongo.Database) *TestFlowRepo {
	return &TestFlowRepo{caseCol: db.Collection("testflows")}
}

func (c *TestFlowRepo) List(ctx context.Context, req *types.PageReq[ListTestFlowParams]) (*types.PageResp[types.TestFlow], error) {
	filter := bson.D{}
	if !req.Params.GroupID.IsZero() {
		filter = append(filter, bson.E{Key: "groupId", Value: req.Params.GroupID})
	}

	count, err := c.caseCol.CountDocuments(ctx, filter)
	if err != nil {
		return nil, err
	}

	cur, err := c.caseCol.Find(ctx, filter, sortModifyDesc)
	if err != nil {
		return nil, err
	}

	tf := []types.TestFlow{}
	err = cur.All(ctx, &tf)
	if err != nil {
		return nil, err
	}
	return &types.PageResp[types.TestFlow]{
		Total:   count,
		PageNum: 1,
		Pages:   1,
		List:    tf,
	}, nil
}

func (c *TestFlowRepo) Get(ctx context.Context, params *GetTestFlowParams) (*types.TestFlow, error) {
	filter := bson.M{}
	if len(params.Name) > 0 {
		filter["name"] = params.Name
	}
	if !params.ID.IsZero() {
		filter["_id"] = params.ID
	}

	tf := &types.TestFlow{}
	err := c.caseCol.FindOne(ctx, filter).Decode(tf)
	if err != nil {
		return nil, err
	}
	return tf, nil
}

func (c *TestFlowRepo) CountByGroup(ctx context.Context, groupId primitive.ObjectID) (int64, error) {
	return c.caseCol.CountDocuments(ctx, bson.D{{"groupId", groupId}})
}

func (c *TestFlowRepo) Save(ctx context.Context, tf types.TestFlow) (primitive.ObjectID, error) {
	if tf.ID.IsZero() {
		tf.ID = primitive.NewObjectID()
	}

	count, err := c.caseCol.CountDocuments(ctx, bson.D{{"_id", tf.ID}})
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
	_, err = c.caseCol.UpdateOne(ctx, bson.D{{"name", tf.Name}}, update, options.Update().SetUpsert(true))
	if err != nil {
		return primitive.ObjectID{}, err
	}

	return tf.ID, nil
}

func (c *TestFlowRepo) Delete(ctx context.Context, id primitive.ObjectID) error {
	tf := &types.TestFlow{}
	err := c.caseCol.FindOne(ctx, bson.D{{"_id", id}}).Decode(tf)
	if err != nil || err == mongo.ErrNoDocuments {
		return fmt.Errorf("test flow (%d) not exis", id)
	}

	_, err = c.caseCol.DeleteOne(ctx, bson.D{{"_id", id}})
	if err != nil {
		return err
	}
	return nil
}

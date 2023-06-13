package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/hunjixin/brightbird/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

type CountGroupParams struct {
	ID   primitive.ObjectID
	Name *string
}

type IGroupRepo interface {
	Count(context.Context, *CountGroupParams) (int64, error)
	List(context.Context) ([]*models.Group, error)
	Get(context.Context, primitive.ObjectID) (*models.Group, error)
	Save(context.Context, models.Group) (primitive.ObjectID, error)
	Delete(ctx context.Context, id primitive.ObjectID) error
}

var _ IGroupRepo = (*GroupSvc)(nil)

type GroupSvc struct {
	groupCol    *mongo.Collection
	testflowSvc ITestFlowRepo
}

func NewGroupSvc(ctx context.Context, db *mongo.Database, testflowSvc ITestFlowRepo) (*GroupSvc, error) {
	col := db.Collection("groups")
	_, err := col.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bsonx.Doc{{Key: "name", Value: bsonx.Int32(-1)}},
		},
	})
	if err != nil {
		return nil, err
	}
	return &GroupSvc{groupCol: col, testflowSvc: testflowSvc}, nil
}

func (g *GroupSvc) List(ctx context.Context) ([]*models.Group, error) {
	cur, err := g.groupCol.Find(ctx, bson.M{}, sortModifyDesc)
	if err != nil {
		return nil, err
	}

	groups := []*models.Group{}
	err = cur.All(ctx, &groups)
	if err != nil {
		return nil, err
	}
	return groups, nil
}

func (g *GroupSvc) Count(ctx context.Context, params *CountGroupParams) (int64, error) {
	filter := bson.M{}
	if params.Name != nil {
		filter["name"] = params.Name
	}
	if !params.ID.IsZero() {
		filter["_id"] = params.ID
	}
	return g.groupCol.CountDocuments(ctx, filter)
}

func (g *GroupSvc) Get(ctx context.Context, id primitive.ObjectID) (*models.Group, error) {
	tf := &models.Group{}
	err := g.groupCol.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(tf)
	if err != nil {
		return nil, err
	}
	return tf, nil
}

func (g *GroupSvc) Delete(ctx context.Context, id primitive.ObjectID) error {
	count, err := g.testflowSvc.Count(ctx, &CountTestFlowParams{
		GroupID: id,
	})
	if err != nil {
		return err
	}

	if count > 0 {
		return fmt.Errorf("test flow (%d) in this group, remove test flow first", count)
	}

	_, err = g.groupCol.DeleteOne(ctx, bson.D{{Key: "_id", Value: id}})
	if err != nil {
		return err
	}
	return nil
}

func (g *GroupSvc) Save(ctx context.Context, group models.Group) (primitive.ObjectID, error) {
	if group.ID.IsZero() {
		group.ID = primitive.NewObjectID()
	}

	count, err := g.groupCol.CountDocuments(ctx, bson.D{{Key: "_id", Value: group.ID}})
	if err != nil {
		return primitive.ObjectID{}, err
	}
	if count == 0 {
		group.BaseTime.CreateTime = time.Now().Unix()
		group.BaseTime.ModifiedTime = time.Now().Unix()
	} else {
		group.BaseTime.ModifiedTime = time.Now().Unix()
	}

	update := bson.M{
		"$set": group,
	}
	_, err = g.groupCol.UpdateOne(ctx, bson.D{{Key: "name", Value: group.Name}}, update, options.Update().SetUpsert(true))
	if err != nil {
		return primitive.ObjectID{}, err
	}
	return group.ID, nil
}

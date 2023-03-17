package services

import (
	"context"
	"fmt"
	"github.com/hunjixin/brightbird/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IGroupService interface {
	List(context.Context) ([]*types.Group, error)
	Get(context.Context, primitive.ObjectID) (*types.Group, error)
	Save(context.Context, types.Group) error
	Delete(ctx context.Context, id primitive.ObjectID) error
}

var _ IGroupService = (*GroupSvc)(nil)

type GroupSvc struct {
	groupCol    *mongo.Collection
	testflowSvc ITestFlowService
}

func NewGroupSvc(groupCol *mongo.Collection, testflowSvc ITestFlowService) *GroupSvc {
	return &GroupSvc{groupCol: groupCol, testflowSvc: testflowSvc}
}

func (g *GroupSvc) List(ctx context.Context) ([]*types.Group, error) {
	cur, err := g.groupCol.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	var tf []*types.Group
	err = cur.All(ctx, &tf)
	if err != nil {
		return nil, err
	}
	return tf, nil
}

func (g GroupSvc) Get(ctx context.Context, id primitive.ObjectID) (*types.Group, error) {
	tf := &types.Group{}
	err := g.groupCol.FindOne(ctx, bson.D{{"_id", id}}).Decode(tf)
	if err != nil {
		return nil, err
	}
	return tf, nil
}

func (g GroupSvc) Delete(ctx context.Context, id primitive.ObjectID) error {
	count, err := g.testflowSvc.CountByGroup(ctx, id)
	if count > 0 {
		return fmt.Errorf("test flow (%d) in this group, remove test flow first", count)
	}
	_, err = g.groupCol.DeleteOne(ctx, bson.D{{"_id", id}})
	if err != nil {
		return err
	}
	return nil
}

func (g *GroupSvc) Save(ctx context.Context, group types.Group) error {
	if group.ID.IsZero() {
		group.ID = primitive.NewObjectID()
	}
	update := bson.M{
		"$set": group,
	}
	_, err := g.groupCol.UpdateOne(ctx, bson.D{{"Name", group.Name}}, update, options.Update().SetUpsert(true))
	if err != nil {
		return err
	}
	return nil
}

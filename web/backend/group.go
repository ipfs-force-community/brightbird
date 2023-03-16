package main

import (
	"context"
	"github.com/hunjixin/brightbird/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IGroupService interface {
	List(context.Context) ([]*types.Group, error)
	Get(context.Context, string) (*types.Group, error)
	Save(context.Context, types.Group) error
}

var _ IGroupService = (*GroupSvc)(nil)

type GroupSvc struct {
	col *mongo.Collection
}

func (g *GroupSvc) List(ctx context.Context) ([]*types.Group, error) {
	cur, err := g.col.Find(ctx, bson.M{})
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

func (g GroupSvc) Get(ctx context.Context, s string) (*types.Group, error) {
	tf := &types.Group{}
	err := g.col.FindOne(ctx, bson.D{{"Name", s}}).Decode(tf)
	if err != nil {
		return nil, err
	}
	return tf, nil
}

func (g *GroupSvc) Save(ctx context.Context, group types.Group) error {
	if group.ID.IsZero() {
		group.ID = primitive.NewObjectID()
	}
	update := bson.M{
		"$set": group,
	}
	_, err := g.col.UpdateOne(ctx, bson.D{{"Name", group.Name}}, update, options.Update().SetUpsert(true))
	if err != nil {
		return err
	}
	return nil
}

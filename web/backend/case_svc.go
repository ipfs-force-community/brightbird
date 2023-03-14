package main

import (
	"context"
	"github.com/hunjixin/brightbird/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ITestCaseService interface {
	Get(ctx context.Context, name string) (*types.TestFlow, error)
	List(ctx context.Context) ([]*types.TestFlow, error)
	Plugins(ctx context.Context) ([]PluginOut, error)
	Save(ctx context.Context, testcase types.TestFlow) error
}

type CaseSvc struct {
	caseCol         *mongo.Collection
	execPluginStore ExecPluginStore
}

func (c *CaseSvc) List(ctx context.Context) ([]*types.TestFlow, error) {
	cur, err := c.caseCol.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	var tf []*types.TestFlow
	err = cur.All(ctx, &tf)
	if err != nil {
		return nil, err
	}
	return tf, nil
}

func (c *CaseSvc) Get(ctx context.Context, name string) (*types.TestFlow, error) {
	tf := &types.TestFlow{}
	err := c.caseCol.FindOne(ctx, bson.D{{"Name", name}}).Decode(tf)
	if err != nil {
		return nil, err
	}
	return tf, nil
}

func (c *CaseSvc) Plugins(ctx context.Context) ([]PluginOut, error) {
	var deployPlugins []PluginOut
	err := c.execPluginStore.Each(func(detail *types.PluginDetail) error {
		pluginOut, err := getPluginOutput(detail)
		if err != nil {
			return err
		}
		deployPlugins = append(deployPlugins, pluginOut)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return deployPlugins, nil
}

func (c *CaseSvc) Save(ctx context.Context, tf types.TestFlow) error {
	update := bson.M{
		"$set": tf,
	}
	_, err := c.caseCol.UpdateOne(ctx, bson.D{{"Name", tf.Name}}, update, options.Update().SetUpsert(true))
	if err != nil {
		return err
	}
	return nil
}

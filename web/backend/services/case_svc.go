package services

import (
	"context"
	"github.com/hunjixin/brightbird/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ITestFlowService interface {
	GetByName(context.Context, string) (*types.TestFlow, error)
	GetById(context.Context, primitive.ObjectID) (*types.TestFlow, error)
	List(context.Context) (*types.PageResp[types.TestFlow], error)
	Plugins(context.Context) ([]types.PluginOut, error)
	Save(context.Context, types.TestFlow) error
	CountByGroup(ctx context.Context, groupId primitive.ObjectID) (int64, error)
	ListInGroup(context.Context, *types.PageReq[string]) (*types.PageResp[types.TestFlow], error)
}

type CaseSvc struct {
	caseCol         *mongo.Collection
	execPluginStore ExecPluginStore
}

func NewCaseSvc(caseCol *mongo.Collection, execPluginStore ExecPluginStore) *CaseSvc {
	return &CaseSvc{caseCol: caseCol, execPluginStore: execPluginStore}
}

type BasePage struct {
	Total   int `json:"total"`
	Pages   int `json:"pages"`
	PageNum int `json:"pageNum"`
}

func (c *CaseSvc) List(ctx context.Context) (*types.PageResp[types.TestFlow], error) {
	cur, err := c.caseCol.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	tf := []types.TestFlow{}
	err = cur.All(ctx, &tf)
	if err != nil {
		return nil, err
	}
	return &types.PageResp[types.TestFlow]{
		Total:   len(tf),
		PageNum: 1,
		Pages:   1,
		List:    tf,
	}, nil
}

func (c *CaseSvc) ListInGroup(ctx context.Context, req *types.PageReq[string]) (*types.PageResp[types.TestFlow], error) {
	groupId, err := primitive.ObjectIDFromHex(req.Params)
	if err != nil {
		return nil, err
	}

	cur, err := c.caseCol.Find(ctx, bson.M{"groupId": groupId})
	if err != nil {
		return nil, err
	}

	tf := []types.TestFlow{}
	err = cur.All(ctx, &tf)
	if err != nil {
		return nil, err
	}
	return &types.PageResp[types.TestFlow]{
		Total:   len(tf),
		PageNum: 1,
		Pages:   1,
		List:    tf,
	}, nil
}

func (c *CaseSvc) GetByName(ctx context.Context, name string) (*types.TestFlow, error) {
	tf := &types.TestFlow{}
	err := c.caseCol.FindOne(ctx, bson.D{{"Name", name}}).Decode(tf)
	if err != nil {
		return nil, err
	}
	return tf, nil
}

func (c *CaseSvc) GetById(ctx context.Context, id primitive.ObjectID) (*types.TestFlow, error) {
	tf := &types.TestFlow{}
	err := c.caseCol.FindOne(ctx, bson.D{{"_id", id}}).Decode(tf)
	if err != nil {
		return nil, err
	}
	return tf, nil
}

func (c *CaseSvc) Plugins(ctx context.Context) ([]types.PluginOut, error) {
	var deployPlugins []types.PluginOut
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

func (c *CaseSvc) CountByGroup(ctx context.Context, groupId primitive.ObjectID) (int64, error) {
	return c.caseCol.CountDocuments(ctx, bson.D{{"groupId", groupId}})
}
func (c *CaseSvc) Save(ctx context.Context, tf types.TestFlow) error {
	if tf.ID.IsZero() {
		tf.ID = primitive.NewObjectID()
	}

	update := bson.M{
		"$set": tf,
	}
	_, err := c.caseCol.UpdateOne(ctx, bson.D{{"name", tf.Name}}, update, options.Update().SetUpsert(true))
	if err != nil {
		return err
	}
	return nil
}

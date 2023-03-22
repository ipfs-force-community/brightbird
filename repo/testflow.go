package repo

import (
	"context"
	"sort"
	"time"

	"github.com/hunjixin/brightbird/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ITestFlowRepo interface {
	GetByName(context.Context, string) (*types.TestFlow, error)
	GetById(context.Context, primitive.ObjectID) (*types.TestFlow, error)
	List(context.Context) (*types.PageResp[types.TestFlow], error)
	Plugins(context.Context) ([]types.PluginOut, error)
	Save(context.Context, types.TestFlow) (primitive.ObjectID, error)
	CountByGroup(ctx context.Context, groupId primitive.ObjectID) (int64, error)
	ListInGroup(context.Context, *types.PageReq[string]) (*types.PageResp[types.TestFlow], error)
}

type TestFlowRepo struct {
	caseCol         *mongo.Collection
	execPluginStore ExecPluginStore
}

func NewTestFlowRepo(db *mongo.Database, execPluginStore ExecPluginStore) *TestFlowRepo {
	return &TestFlowRepo{caseCol: db.Collection("testflows"), execPluginStore: execPluginStore}
}

type BasePage struct {
	Total   int `json:"total"`
	Pages   int `json:"pages"`
	PageNum int `json:"pageNum"`
}

func (c *TestFlowRepo) List(ctx context.Context) (*types.PageResp[types.TestFlow], error) {
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

func (c *TestFlowRepo) ListInGroup(ctx context.Context, req *types.PageReq[string]) (*types.PageResp[types.TestFlow], error) {
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

func (c *TestFlowRepo) GetByName(ctx context.Context, name string) (*types.TestFlow, error) {
	tf := &types.TestFlow{}
	err := c.caseCol.FindOne(ctx, bson.D{{"name", name}}).Decode(tf)
	if err != nil {
		return nil, err
	}
	return tf, nil
}

func (c *TestFlowRepo) GetById(ctx context.Context, id primitive.ObjectID) (*types.TestFlow, error) {
	tf := &types.TestFlow{}
	err := c.caseCol.FindOne(ctx, bson.D{{"_id", id}}).Decode(tf)
	if err != nil {
		return nil, err
	}
	return tf, nil
}

func (c *TestFlowRepo) Plugins(ctx context.Context) ([]types.PluginOut, error) {
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
	sort.Slice(deployPlugins, func(i, j int) bool {
		return deployPlugins[i].Name > deployPlugins[j].Name
	})
	return deployPlugins, nil
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

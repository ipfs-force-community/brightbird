package repo

import (
	"context"
	"fmt"

	"github.com/hunjixin/brightbird/types"

	"github.com/hunjixin/brightbird/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

type IPluginService interface {
	DeployPlugins(context.Context) ([]models.PluginOut, error)
	ExecPlugins(context.Context) ([]models.PluginOut, error)
	GetPlugin(context.Context, string, string) (*models.PluginOut, error)
	SavePlugins(context.Context, []*models.PluginOut) error
}

type PluginSvc struct {
	pluginCol *mongo.Collection
}

func NewPluginSvc(ctx context.Context, db *mongo.Database) (*PluginSvc, error) {
	col := db.Collection("plugins")
	_, err := col.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bsonx.Doc{
				{Key: "name", Value: bsonx.Int32(-1)},
				{Key: "version", Value: bsonx.Int32(-1)},
			},
			Options: options.Index().SetUnique(true),
		},
	})
	if err != nil {
		return nil, err
	}
	return &PluginSvc{
		pluginCol: col,
	}, nil
}

func (p *PluginSvc) DeployPlugins(ctx context.Context) ([]models.PluginOut, error) {
	var pluginOut []models.PluginOut
	cur, err := p.pluginCol.Find(ctx, bson.D{{"plugintype", types.Deploy}}, sortNameDesc)
	if err != nil {
		return nil, err
	}
	err = cur.All(ctx, &pluginOut)
	if err != nil {
		return nil, err
	}
	return pluginOut, nil
}

func (p *PluginSvc) ExecPlugins(ctx context.Context) ([]models.PluginOut, error) {
	var pluginOut []models.PluginOut
	cur, err := p.pluginCol.Find(ctx, bson.D{{"plugintype", types.TestExec}}, sortNameDesc)
	if err != nil {
		return nil, err
	}
	err = cur.All(ctx, &pluginOut)
	if err != nil {
		return nil, err
	}
	return pluginOut, nil
}

func (p *PluginSvc) SavePlugins(ctx context.Context, pluginOuts []*models.PluginOut) error {
	//do some check
	for _, plugin := range pluginOuts {
		count, err := p.pluginCol.CountDocuments(ctx, bson.M{"name": plugin.Name, "version": plugin.Version})
		if err != nil {
			return err
		}
		if count > 0 {
			return fmt.Errorf("plugin %s version %s already exit, please remove first", plugin.Name, plugin.Version)
		}
	}

	var models []mongo.WriteModel
	for _, p := range pluginOuts {
		p.ID = primitive.NewObjectID()
		models = append(models, mongo.NewInsertOneModel().SetDocument(p))
	}

	_, err := p.pluginCol.BulkWrite(ctx, models)
	if err != nil {
		return err
	}

	return nil
}

func (p *PluginSvc) GetPlugin(ctx context.Context, name, version string) (*models.PluginOut, error) {
	var pluginOut *models.PluginOut
	err := p.pluginCol.FindOne(ctx, bson.D{{"name", name}, {"version", version}}).Decode(&pluginOut)
	if err != nil {
		return nil, err
	}
	return pluginOut, nil
}

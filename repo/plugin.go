package repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hunjixin/brightbird/models"
	"github.com/hunjixin/brightbird/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

var ErrPluginNotFound = errors.New("plugin not found")

// ListPluginParams
// swagger:parameters listPluginParams
type ListPluginParams struct {
	// name of plugin
	Name *string `form:"name" json:"name"`
	// plugin type
	PluginType *types.PluginType `form:"pluginType" json:"pluginType"`
}

// GetPluginParams
// swagger:parameters getPluginParams
type GetPluginParams struct {
	// id of plugin
	ID *string `form:"id" json:"id"`
	// name of plugin
	Name *string `form:"name" json:"name"`
	// plugin type
	PluginType *types.PluginType `form:"pluginType" json:"pluginType"`
}

// AddLabelParams
// swagger:parameters addLabelParams
type AddLabelParams struct {
	// name of plugin
	Name *string `form:"name" json:"name"`
	// plugin type
	Label *string `form:"label" json:"label"`
}

// DeleteLabelParams
// swagger:parameters deleteLabelParams
type DeleteLabelParams struct {
	// name of plugin
	Name *string `form:"name" json:"name"`
	// plugin type
	Label *string `form:"label" json:"label"`
}

// DeletePluginParams
type DeletePluginParams struct {
	// id of plugin
	ID primitive.ObjectID
	// specific plugin version
	Version string
}

func NewListPluginParams() *ListPluginParams {
	return &ListPluginParams{}
}

func (params *ListPluginParams) SetPluginType(pluginType types.PluginType) *ListPluginParams {
	params.PluginType = &pluginType
	return params
}

func (params *ListPluginParams) SetName(name string) *ListPluginParams {
	params.Name = &name
	return params
}

// ListMainFestParams
// swagger:parameters listMainFestParams
type ListMainFestParams struct {
	Name       *string           `form:"name" json:"name"`
	PluginType *types.PluginType `form:"pluginType" json:"pluginType"`
}

func NewListMainFestParams() *ListMainFestParams {
	return &ListMainFestParams{}
}

func (params *ListMainFestParams) SetPluginType(pluginType types.PluginType) *ListMainFestParams {
	params.PluginType = &pluginType
	return params
}

func (params *ListMainFestParams) SetName(name string) *ListMainFestParams {
	params.Name = &name
	return params
}

type IPluginService interface {
	AddLabel(context.Context, string, string) error
	DeleteLabel(context.Context, string, string) error
	DeletePluginByVersion(context.Context, *DeletePluginParams) error
	GetPluginDetail(context.Context, *GetPluginParams) (*models.PluginDetail, error)
	ListPlugin(context.Context, *ListPluginParams) ([]*models.PluginDetail, error)
	GetPlugin(context.Context, string, string) (*models.Plugin, error)
	SavePlugins(context.Context, *models.Plugin) error
	CountPlugin(ctx context.Context, pluginType *types.PluginType) (int64, error)
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

func (p *PluginSvc) ListPlugin(ctx context.Context, listPluginParams *ListPluginParams) ([]*models.PluginDetail, error) {
	filter := bson.D{}
	if listPluginParams.Name != nil {
		filter = append(filter, bson.E{Key: "name", Value: listPluginParams.Name})
	}

	if listPluginParams.PluginType != nil {
		filter = append(filter, bson.E{Key: "plugintype", Value: listPluginParams.PluginType})
	}

	cur, err := p.pluginCol.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	var plugins []*models.PluginDetail
	err = cur.All(ctx, &plugins)
	if err != nil {
		return nil, err
	}
	return plugins, nil
}

func (p *PluginSvc) GetPluginDetail(ctx context.Context, getPluginParams *GetPluginParams) (*models.PluginDetail, error) {
	filter := bson.D{}
	if getPluginParams.ID != nil {
		id, err := primitive.ObjectIDFromHex(*getPluginParams.ID)
		if err != nil {
			return nil, err
		}
		filter = append(filter, bson.E{Key: "_id", Value: id})
	}

	if getPluginParams.Name != nil {
		filter = append(filter, bson.E{Key: "name", Value: getPluginParams.Name})
	}

	if getPluginParams.PluginType != nil {
		filter = append(filter, bson.E{Key: "plugintype", Value: getPluginParams.PluginType})
	}

	var plugin models.PluginDetail
	err := p.pluginCol.FindOne(ctx, filter).Decode(&plugin)
	if err != nil {
		return nil, err
	}

	return &plugin, nil
}

func (p *PluginSvc) DeletePluginByVersion(ctx context.Context, params *DeletePluginParams) error {
	update := bson.M{
		"$pull": bson.M{
			"plugins": bson.M{
				"version": params.Version,
			},
		},
	}

	_, err := p.pluginCol.UpdateOne(ctx, bson.D{{Key: "_id", Value: params.ID}}, update)
	if err != nil {
		return err
	}

	var plugin models.PluginDetail
	err = p.pluginCol.FindOne(ctx, bson.M{"_id": params.ID}).Decode(&plugin)
	if err != nil {
		return err
	}

	if len(plugin.Plugins) == 0 {
		_, err = p.pluginCol.DeleteOne(ctx, bson.M{"_id": params.ID})
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *PluginSvc) SavePlugins(ctx context.Context, pluginImport *models.Plugin) error {
	//do some check
	pluginDetail := &models.PluginDetail{}
	err := p.pluginCol.FindOne(ctx, bson.M{"name": pluginImport.Name}).Decode(pluginDetail)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			//add a plugin collection
			pluginDetail.ID = primitive.NewObjectID()
			pluginDetail = &models.PluginDetail{
				ID:          primitive.NewObjectID(),
				Name:        pluginImport.Name,
				PluginType:  pluginImport.PluginType,
				Description: pluginImport.Description,
				Labels:      []string{pluginImport.Name}, //set name as default label
				Plugins:     []models.Plugin{*pluginImport},
				BaseTime: models.BaseTime{
					CreateTime:   time.Now().Unix(),
					ModifiedTime: time.Now().Unix(),
				},
			}
			_, err = p.pluginCol.InsertOne(ctx, pluginDetail)
			return err
		}
		return err
	}
	//insert a version
	count, err := p.pluginCol.CountDocuments(ctx, bson.M{"name": pluginImport.Name, "plugins": bson.M{"$elemMatch": bson.M{
		"version": pluginImport.Version,
	}}})
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("plugin %s version %s already exit, please remove first", pluginImport.Name, pluginImport.Version)
	}
	if pluginImport.PluginType != pluginDetail.PluginType {
		return fmt.Errorf("plugin type change")
	}

	update := bson.M{
		"$set": bson.M{
			"description":  pluginImport.Description,
			"modifiedtime": time.Now().Unix(),
		},
		"$push": bson.M{
			"plugins": pluginImport,
		},
	}
	_, err = p.pluginCol.UpdateOne(ctx, bson.M{"name": pluginImport.Name}, update)
	return err
}

func (p *PluginSvc) GetPlugin(ctx context.Context, name, version string) (*models.Plugin, error) {
	plugin := &models.PluginDetail{}
	err := p.pluginCol.FindOne(ctx, bson.M{"name": name}).Decode(plugin)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("plugin %s(%s) %w", name, version, ErrPluginNotFound)
		}
		return nil, err
	}
	for _, p := range plugin.Plugins {
		if p.Version == version {
			return &p, nil
		}
	}
	return nil, fmt.Errorf("plugin %s(%s) %w", name, version, ErrPluginNotFound)
}

func (p *PluginSvc) AddLabel(ctx context.Context, name string, newLabel string) error {
	count, err := p.pluginCol.CountDocuments(ctx, bson.M{"name": name, "labels": newLabel})
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("plugin %s label %s already exit, please remove first", name, newLabel)
	}

	_, err = p.pluginCol.UpdateOne(ctx, bson.M{"name": name}, bson.M{"$push": bson.M{
		"labels": newLabel,
	}})
	return err
}

func (p *PluginSvc) DeleteLabel(ctx context.Context, name string, toDeleteLabel string) error {
	count, err := p.pluginCol.CountDocuments(ctx, bson.M{"name": name, "labels": toDeleteLabel})
	if err != nil {
		return err
	}
	if count == 0 {
		return nil
	}

	_, err = p.pluginCol.UpdateOne(ctx, bson.M{"name": name}, bson.M{"$pull": bson.M{
		"labels": toDeleteLabel,
	}})
	return err
}

func (p *PluginSvc) CountPlugin(ctx context.Context, pluginType *types.PluginType) (int64, error) {
	filter := bson.M{}
	filter["plugintype"] = *pluginType

	count, err := p.pluginCol.CountDocuments(ctx, filter)
	if err != nil {
		return -1, err
	}
	return count, nil
}

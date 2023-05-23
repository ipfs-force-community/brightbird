package repo

import (
	"context"
	"fmt"
	"sort"

	"github.com/hunjixin/brightbird/models"
	"github.com/hunjixin/brightbird/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

// ListPluginParams
// swagger:model listPluginParams
type ListPluginParams struct {
	Name       *string           `form:"name"`
	Version    *string           `form:"version"`
	PluginType *types.PluginType `form:"pluginType"`
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

func (params *ListPluginParams) SetVersion(version string) *ListPluginParams {
	params.Version = &version
	return params
}

// ListMainFestParams
// swagger:model listMainFestParams
type ListMainFestParams struct {
	Name       *string           `form:"name"`
	PluginType *types.PluginType `form:"pluginType"`
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
	DeletePlugin(context.Context, primitive.ObjectID) error
	ListPlugin(context.Context, *ListPluginParams) ([]*models.PluginDetail, error)
	GetPlugin(context.Context, string, string) (*models.PluginDetail, error)
	SavePlugins(context.Context, []*models.PluginDetail) error
	PluginSummary(context.Context, *ListMainFestParams) ([]*models.PluginInfo, error)
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
		filter = append(filter, bson.E{"name", listPluginParams.Name})
	}

	if listPluginParams.Version != nil {
		filter = append(filter, bson.E{"version", listPluginParams.Version})
	}

	if listPluginParams.PluginType != nil {
		filter = append(filter, bson.E{"plugintype", listPluginParams.PluginType})
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
func (p *PluginSvc) DeletePlugin(ctx context.Context, id primitive.ObjectID) error {
	_, err := p.pluginCol.DeleteMany(ctx, bson.D{{"_id", id}})
	return err
}

func (p *PluginSvc) SavePlugins(ctx context.Context, pluginOuts []*models.PluginDetail) error {
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

func (p *PluginSvc) GetPlugin(ctx context.Context, name, version string) (*models.PluginDetail, error) {
	var pluginDetail *models.PluginDetail
	err := p.pluginCol.FindOne(ctx, bson.D{{"name", name}, {"version", version}}).Decode(&pluginDetail)
	if err != nil {
		return nil, err
	}
	return pluginDetail, nil
}

func (p *PluginSvc) PluginSummary(ctx context.Context, listPluginParams *ListMainFestParams) ([]*models.PluginInfo, error) {
	filter := bson.D{}
	if listPluginParams.Name != nil {
		filter = append(filter, bson.E{"name", listPluginParams.Name})
	}

	if listPluginParams.PluginType != nil {
		filter = append(filter, bson.E{"plugintype", listPluginParams.PluginType})
	}

	matchStage := bson.D{{"$match", filter}}
	groupStage := bson.D{
		{"$group", bson.D{
			{"_id", bson.D{
				{"name", "$name"},
				{"plugintype", "$plugintype"},
			}},
			{"description", bson.M{"$last": "$description"}},
		},
		},
	}
	projectStage := bson.D{
		{"$project", bson.D{
			{"_id", 0},
			{"description", 1},
			{"name", "$_id.name"},
			{"plugintype", "$_id.plugintype"},
		}},
	}
	groupResultsCur, err := p.pluginCol.Aggregate(ctx, mongo.Pipeline{matchStage, groupStage, projectStage}, options.Aggregate().SetAllowDiskUse(true))
	if err != nil {
		return nil, err
	}

	var pluginSummary []*models.PluginInfo
	err = groupResultsCur.All(ctx, &pluginSummary)
	if err != nil {
		return nil, err
	}
	sort.Slice(pluginSummary, func(i, j int) bool {
		return pluginSummary[i].Name < pluginSummary[j].Name
	})
	return pluginSummary, nil
}

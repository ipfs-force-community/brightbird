package main

import (
	"github.com/hunjixin/brightbird/web/backend/services"
	"go.mongodb.org/mongo-driver/mongo"
)

func NewCaseSvc(db *mongo.Database, execPluginStore services.ExecPluginStore) services.ITestFlowService {
	return services.NewCaseSvc(db.Collection("cases"), execPluginStore)
}

func NewGroupSvc(db *mongo.Database, testflowSvc services.ITestFlowService) services.IGroupService {
	return services.NewGroupSvc(db.Collection("group"), testflowSvc)
}

func NewPlugin(deployPluginStore services.DeployPluginStore) services.IPluginService {
	return services.NewPluginSvc(deployPluginStore)
}

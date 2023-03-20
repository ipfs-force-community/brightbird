package main

import (
	"github.com/gin-gonic/gin"
	"github.com/hunjixin/brightbird/types"
	"go.mongodb.org/mongo-driver/mongo"
)

type DeployPluginStore types.IPluginInfo

type ExecPluginStore types.IPluginInfo

type V1RouterGroup = gin.RouterGroup

func NewCaseSvc(db *mongo.Database, execPluginStore ExecPluginStore) ITestCaseService {
	return &CaseSvc{
		caseCol:         db.Collection("cases"),
		execPluginStore: execPluginStore,
	}
}

func NewGroupSvc(db *mongo.Database, execPluginStore ExecPluginStore) IGroupService {
	return &GroupSvc{
		groupCol: db.Collection("group"),
		//testFlowCol: db.Collection("cases"),
	}
}

func NewPlugin(deployPluginStore DeployPluginStore) IPluginService {
	return &PluginSvc{
		deployPluginStore: deployPluginStore,
	}
}

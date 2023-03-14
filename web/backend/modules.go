package main

import (
	"github.com/gin-gonic/gin"
	"github.com/hunjixin/brightbird/types"
	"go.mongodb.org/mongo-driver/mongo"
)

type DeployPluginStore types.IPluginInfo

type ExecPluginStore types.IPluginInfo

type V1RouterGroup = gin.RouterGroup

func NewCaseSvc(caseCol *mongo.Collection, execPluginStore ExecPluginStore) ITestCaseService {
	return &CaseSvc{
		caseCol:         caseCol,
		execPluginStore: execPluginStore,
	}
}

func NewPlugin(deployPluginStore DeployPluginStore) IPluginService {
	return &PluginSvc{
		deployPluginStore: deployPluginStore,
	}
}

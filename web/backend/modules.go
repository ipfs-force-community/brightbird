package main

import (
	"github.com/hunjixin/brightbird/repo"
	"go.mongodb.org/mongo-driver/mongo"
)

func NewTestFlowRepo(db *mongo.Database, execPluginStore repo.ExecPluginStore) repo.ITestFlowRepo {
	return repo.NewTestFlowRepo(db, execPluginStore)
}

func NewGroupRepo(db *mongo.Database, testflowSvc repo.ITestFlowRepo) repo.IGroupRepo {
	return repo.NewGroupSvc(db, testflowSvc)
}

func NewJobRepo(db *mongo.Database) repo.IJobRepo {
	return repo.NewJobRepo(db)
}

func NewTaskRepo(db *mongo.Database) repo.ITaskRepo {
	return repo.NewTaskRepo(db)
}

func NewPlugin(deployPluginStore repo.DeployPluginStore) repo.IPluginService {
	return repo.NewPluginSvc(deployPluginStore)
}

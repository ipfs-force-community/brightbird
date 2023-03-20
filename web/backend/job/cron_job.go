package job

import (
	"context"
	"github.com/google/uuid"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/repo"
	"github.com/hunjixin/brightbird/types"
	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var _ IJob = (*CronJob)(nil)

type CronJob struct {
	job          types.Job
	c            *cron.Cron
	taskRepo     repo.ITaskRepo
	testFlowRepo repo.ITestFlowRepo
	k8sEnv       *env.K8sEnvDeployer
	imageBuilder ImageBuilderMgr
}

func (cronJob *CronJob) Id() string {
	return cronJob.job.ID.String()
}

func (cronJob *CronJob) Run(ctx context.Context) error {
	log := log.With("job", cronJob.job.ID, "testflow", cronJob.job.TestFlowId)
	_, err := cronJob.c.AddFunc(cronJob.job.CronExpression, func() {
		log.Infof("job(%s) start to running", cronJob.job.Name)
		err := cronJob.taskRepo.Save(ctx, types.Task{
			ID:         primitive.NewObjectID(),
			JobId:      cronJob.job.ID,
			TestFlowId: cronJob.job.TestFlowId,
			State:      types.Init,
			TestId:     types.TestId(uuid.New().String()[:8]),
			BaseTime:   types.BaseTime{},
		})
		if err != nil {
			log.Infof("job not running")
		}
	})
	return err
}

func (cronJob *CronJob) Stop(_ context.Context) error {
	select {
	case <-cronJob.c.Stop().Done():
		return nil
	}
}

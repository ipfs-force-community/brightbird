package job

import (
	"context"
	"github.com/google/uuid"
	"github.com/hunjixin/brightbird/repo"
	"github.com/hunjixin/brightbird/types"
	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var _ IJob = (*CronJob)(nil)

type CronJob struct {
	job      types.Job
	cron     *cron.Cron
	taskRepo repo.ITaskRepo

	cronId *cron.EntryID
}

func NewCronJob(job types.Job, cron *cron.Cron, taskRepo repo.ITaskRepo) *CronJob {
	return &CronJob{job: job, cron: cron, taskRepo: taskRepo}
}

func (cronJob *CronJob) Id() string {
	return cronJob.job.ID.String()
}

func (cronJob *CronJob) Run(ctx context.Context) error {
	log := log.With("job", cronJob.job.ID, "testflow", cronJob.job.TestFlowId)
	entryId, err := cronJob.cron.AddFunc(cronJob.job.CronExpression, func() {
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
	cronJob.cronId = &entryId
	return err
}

func (cronJob *CronJob) Stop(_ context.Context) error {
	if cronJob.cronId != nil {
		cronJob.cron.Remove(*cronJob.cronId)
	}
	return nil
}

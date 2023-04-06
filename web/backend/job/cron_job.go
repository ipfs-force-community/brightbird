package job

import (
	"context"
	"strconv"

	"github.com/google/uuid"
	"github.com/hunjixin/brightbird/repo"
	"github.com/hunjixin/brightbird/types"
	logging "github.com/ipfs/go-log/v2"
	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var cronLog = logging.Logger("cron_job")

var _ IJob = (*CronJob)(nil)

type CronJob struct {
	job      types.Job
	cron     *cron.Cron
	taskRepo repo.ITaskRepo
	jobRepo  repo.IJobRepo

	cronId *cron.EntryID
}

func NewCronJob(job types.Job, cron *cron.Cron, taskRepo repo.ITaskRepo, jobRepo repo.IJobRepo) *CronJob {
	return &CronJob{job: job, cron: cron, taskRepo: taskRepo, jobRepo: jobRepo}
}

func (cronJob *CronJob) Id() string {
	return cronJob.job.ID.String()
}

func (cronJob *CronJob) Run(ctx context.Context) error {
	thisLog := cronLog.With("job", cronJob.job.ID, "testflow", cronJob.job.TestFlowId)
	entryId, err := cronJob.cron.AddFunc(cronJob.job.CronExpression, func() {
		thisLog.Infof("job(%s) start to running", cronJob.job.Name)

		newJob, err := cronJob.jobRepo.IncExecCount(ctx, cronJob.job.ID)
		if err != nil {
			thisLog.Errorf("increase job %s exec count fail %w", cronJob.job.ID, err)
			return
		}

		id, err := cronJob.taskRepo.Save(ctx, &types.Task{
			ID:         primitive.NewObjectID(),
			Name:       cronJob.job.Name + "-" + strconv.Itoa(newJob.ExecCount),
			JobId:      cronJob.job.ID,
			TestFlowId: cronJob.job.TestFlowId,
			State:      types.Init,
			TestId:     types.TestId(uuid.New().String()[:8]),
			BaseTime:   types.BaseTime{},
		})
		if err != nil {
			thisLog.Errorf("job %s save task fail %w", cronJob.job.ID, err)
			return
		}

		thisLog.Infof("job %s save task %s", cronJob.job.ID, id)
	})
	cronJob.cronId = &entryId
	if err == nil {
		thisLog.Infof("add job %s entry %d", cronJob.job.Name, entryId)
	}
	return err
}

func (cronJob *CronJob) Stop(_ context.Context) error {
	if cronJob.cronId != nil {
		cronJob.cron.Remove(*cronJob.cronId)
	}
	return nil
}

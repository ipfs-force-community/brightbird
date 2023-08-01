package job

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/ipfs-force-community/brightbird/models"

	"github.com/google/uuid"
	"github.com/ipfs-force-community/brightbird/repo"
	"github.com/ipfs-force-community/brightbird/types"
	logging "github.com/ipfs/go-log/v2"
	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var cronLog = logging.Logger("cron_job")

var _ IJob = (*CronJob)(nil)

type CronJob struct {
	job      models.Job
	cron     *cron.Cron
	taskRepo repo.ITaskRepo
	jobRepo  repo.IJobRepo

	cronId *cron.EntryID
}

func NewCronJob(job models.Job, cron *cron.Cron, taskRepo repo.ITaskRepo, jobRepo repo.IJobRepo) *CronJob {
	return &CronJob{job: job, cron: cron, taskRepo: taskRepo, jobRepo: jobRepo}
}

func (cronJob *CronJob) ID() string {
	return cronJob.job.ID.Hex()
}

func (cronJob *CronJob) RunImmediately(ctx context.Context) (primitive.ObjectID, error) {
	thisLog := cronLog.With("job", cronJob.job.ID, "testflow", cronJob.job.TestFlowId)
	thisLog.Infof("job(%s) start to running", cronJob.job.Name)

	newJob, err := cronJob.jobRepo.IncExecCount(ctx, cronJob.job.ID)
	if err != nil {
		thisLog.Errorf("increase job %s exec count fail %w", cronJob.job.ID, err)
		return primitive.NilObjectID, err
	}

	id, err := cronJob.taskRepo.Save(ctx, &models.Task{
		ID:              primitive.NewObjectID(),
		Name:            cronJob.job.Name + "-" + strconv.Itoa(newJob.ExecCount),
		JobId:           cronJob.job.ID,
		TestFlowId:      cronJob.job.TestFlowId,
		State:           models.Init,
		TestId:          types.TestId(uuid.New().String()[:8]),
		BaseTime:        models.BaseTime{},
		InheritVersions: cronJob.job.Versions,
	})
	if err != nil {
		thisLog.Errorf("job %s save task fail %w", cronJob.job.ID, err)
		return primitive.NilObjectID, err
	}

	thisLog.Infof("job %s save task %s", cronJob.job.ID, id)
	return id, nil
}

func (cronJob *CronJob) Run(ctx context.Context) error {
	thisLog := cronLog.With("job", cronJob.job.ID, "testflow", cronJob.job.TestFlowId)
	entryId, err := cronJob.cron.AddFunc(cronJob.job.CronExpression, func() {
		_, _ = cronJob.RunImmediately(ctx)
	})
	cronJob.cronId = &entryId
	if err == nil {
		thisLog.Infof("add job %s entry %d", cronJob.job.Name, entryId)
	}
	return err
}

func (cronJob *CronJob) NextNSchedule(_ context.Context, n int) ([]time.Time, error) {
	entry := cronJob.cron.Entry(*cronJob.cronId)
	nextN := []time.Time{entry.Next}
	if n < 1 {
		return nextN, nil
	}

	nextT := entry.Next
	for i := 0; i < n-1; i++ {
		nextT = entry.Schedule.Next(nextT)
		nextN = append(nextN, nextT)
	}
	fmt.Println(nextN)
	return nextN, nil
}

func (cronJob *CronJob) Stop(_ context.Context) error {
	if cronJob.cronId != nil {
		cronJob.cron.Remove(*cronJob.cronId)
	}
	return nil
}

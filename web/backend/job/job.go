package job

import (
	"context"
	"fmt"
	"sync"

	"github.com/hunjixin/brightbird/repo"
	"github.com/hunjixin/brightbird/types"
	"github.com/robfig/cron/v3"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IJobManager interface {
	Start(ctx context.Context) error
	InsertOrReplaceJob(ctx context.Context, job *types.Job) error
	ExecJobImmediately(ctx context.Context, jobId primitive.ObjectID) (primitive.ObjectID, error)
	StopJob(ctx context.Context, jobId primitive.ObjectID) error
	Stop(ctx context.Context) error
}

type IJob interface {
	Id() string
	Run(ctx context.Context) error
	RunImmediately(ctx context.Context) (primitive.ObjectID, error)
	Stop(ctx context.Context) error
}

var _ IJobManager = (*JobManager)(nil)

type JobManager struct {
	lk sync.Mutex

	cron       *cron.Cron
	taskRepo   repo.ITaskRepo
	jobRepo    repo.IJobRepo
	runningJob map[primitive.ObjectID]IJob
}

func NewJobManager(cron *cron.Cron, taskRepo repo.ITaskRepo, jobRepo repo.IJobRepo) *JobManager {
	return &JobManager{
		cron:       cron,
		taskRepo:   taskRepo,
		jobRepo:    jobRepo,
		lk:         sync.Mutex{},
		runningJob: make(map[primitive.ObjectID]IJob),
	}
}

func (j *JobManager) InsertOrReplaceJob(ctx context.Context, job *types.Job) error {
	j.lk.Lock()
	defer j.lk.Unlock()

	oldJob, ok := j.runningJob[job.ID]
	if ok {
		err := oldJob.Stop(ctx)
		if err != nil {
			log.Errorf("unable to stop old job %s %v", job.ID, err)
			return err
		}
		delete(j.runningJob, job.ID)
	}

	switch job.JobType {
	case types.CronJobType:
		jobInstance := NewCronJob(*job, j.cron, j.taskRepo, j.jobRepo)
		err := jobInstance.Run(ctx)
		if err != nil {
			return err
		}
		j.runningJob[job.ID] = jobInstance
	default:
		return fmt.Errorf("unsupport job %s", job.ID)
	}
	return nil
}

func (j *JobManager) Start(ctx context.Context) error {
	jobs, err := j.jobRepo.List(ctx)
	if err != nil {
		return err
	}

	for _, job := range jobs {
		err := j.InsertOrReplaceJob(ctx, job)
		if err != nil {

			log.Info("start job fail %s %v", job.Name, err)
			continue
		}
	}
	j.cron.Start()
	log.Info("start cron job worker")
	return nil
}

func (j *JobManager) ExecJobImmediately(ctx context.Context, jobId primitive.ObjectID) (primitive.ObjectID, error) {
	j.lk.Lock()
	defer j.lk.Unlock()
	if job, ok := j.runningJob[jobId]; ok {
		return job.RunImmediately(ctx)
	} else {
		return primitive.NilObjectID, fmt.Errorf("job %s not running", jobId)
	}
}

func (j *JobManager) StopJob(ctx context.Context, jobId primitive.ObjectID) error {
	j.lk.Lock()
	defer j.lk.Unlock()
	if job, ok := j.runningJob[jobId]; ok {
		err := job.Stop(ctx)
		if err != nil {
			return err
		}
		delete(j.runningJob, jobId)
	}
	return nil
}

func (j *JobManager) Stop(ctx context.Context) error {
	j.lk.Lock()
	defer j.lk.Unlock()
	for jobId, job := range j.runningJob {
		err := job.Stop(ctx)
		if err != nil {
			return err
		}
		delete(j.runningJob, jobId)
	}
	<-j.cron.Stop().Done()
	return nil
}

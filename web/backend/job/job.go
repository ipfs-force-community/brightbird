package job

import (
	"context"
	"sync"

	"github.com/hunjixin/brightbird/repo"
	"github.com/hunjixin/brightbird/types"
	"github.com/robfig/cron/v3"
)

type IJobManager interface {
	Start(ctx context.Context) error
	ReplaceJob(ctx context.Context, job *types.Job) error
	StopJob(ctx context.Context, jobId string) error
}

type IJob interface {
	Id() string
	Run(ctx context.Context) error
	Stop(ctx context.Context) error
}

var _ IJobManager = (*JobManager)(nil)

type JobManager struct {
	lk sync.Mutex

	cron       *cron.Cron
	taskRepo   repo.ITaskRepo
	jobRepo    repo.IJobRepo
	runningJob map[string]IJob
}

func NewJobManager(cron *cron.Cron, taskRepo repo.ITaskRepo, jobRepo repo.IJobRepo) *JobManager {
	return &JobManager{
		cron:       cron,
		taskRepo:   taskRepo,
		jobRepo:    jobRepo,
		lk:         sync.Mutex{},
		runningJob: make(map[string]IJob),
	}
}

func (j *JobManager) ReplaceJob(ctx context.Context, job *types.Job) error {
	j.lk.Lock()
	defer j.lk.Unlock()

	oldJob, ok := j.runningJob[job.ID.String()]
	if ok {
		err := oldJob.Stop(ctx)
		if err != nil {
			log.Errorf("unable to stop old job %s %v", job.ID, err)
			return err
		}
		delete(j.runningJob, job.ID.String())
	}

	switch job.JobType {
	case types.CronJobType:
		jobInstance := NewCronJob(*job, j.cron, j.taskRepo)
		err := jobInstance.Run(ctx)
		if err != nil {
			return err
		}
		j.runningJob[job.ID.String()] = jobInstance
	default:
		log.Errorf("unsupport job %s", job.ID)
	}
	return nil
}

func (j *JobManager) Start(ctx context.Context) error {
	jobs, err := j.jobRepo.List(ctx)
	if err != nil {
		return err
	}

	for _, job := range jobs {
		go func(innerJob *types.Job) {
			switch innerJob.JobType {
			case types.CronJobType:
				jobInstance := NewCronJob(*innerJob, j.cron, j.taskRepo)
				err := jobInstance.Run(ctx)
				if err != nil {
					log.Errorf("job %s unable to start %v", innerJob.ID, err)
					return
				}
				j.lk.Lock()
				defer j.lk.Unlock()
				j.runningJob[innerJob.ID.String()] = jobInstance
			default:
				log.Errorf("unsupport job %s", innerJob.ID)
			}
		}(job)
	}
	j.cron.Start()
	log.Info("start cron job worker")
	return nil
}

func (j *JobManager) StopJob(ctx context.Context, jobId string) error {
	j.lk.Lock()
	defer j.lk.Unlock()
	if job, ok := j.runningJob[jobId]; ok {
		err := job.Stop(ctx)
		if err != nil {
			return err
		}
		delete(j.runningJob, jobId)
	}
	<-j.cron.Stop().Done()
	return nil
}

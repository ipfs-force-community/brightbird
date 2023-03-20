package job

import (
	"context"
	"sync"
)

type JobType string

const (
	ManualJobType  JobType = "manual_job"
	WebHookJobType JobType = "webhook_job"
	CronJobType    JobType = "cron_Job"
)

type IJobManager interface {
	Start(ctx context.Context)
	StopJob(ctx context.Context, jobId string) error
}

type IJob interface {
	Id() string
	Run(ctx context.Context) error
	Stop(ctx context.Context) error
}

var _ IJobManager = (*JobManager)(nil)

type JobManager struct {
	lk   sync.Mutex
	jobs map[string]IJob
}

func (j JobManager) Start(ctx context.Context) {

}

func (j JobManager) StopJob(ctx context.Context, jobId string) error {
	j.lk.Lock()
	defer j.lk.Unlock()
	if job, ok := j.jobs[jobId]; ok {
		err := job.Stop(ctx)
		if err != nil {
			return err
		}
		delete(j.jobs, jobId)
	}
	return nil
}

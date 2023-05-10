package job

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/hunjixin/brightbird/models"

	"github.com/google/go-github/v51/github"
	"github.com/google/uuid"
	"github.com/hunjixin/brightbird/repo"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/web/backend/modules"
	logging "github.com/ipfs/go-log/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

var prMergedjobLog = logging.Logger("tag_create_job")

var _ IJob = (*PRMergedJob)(nil)

type PRMergedJob struct {
	jobId        primitive.ObjectID
	taskRepo     repo.ITaskRepo
	jobRepo      repo.IJobRepo
	testflowRepo repo.ITestFlowRepo

	githubClient *github.Client
	pubsub       modules.WebHookPubsub

	logger *zap.SugaredLogger
}

func NewPRMergedJob(job models.Job, pubsub modules.WebHookPubsub, githubClient *github.Client, taskRepo repo.ITaskRepo, jobRepo repo.IJobRepo, testflowRepo repo.ITestFlowRepo) *PRMergedJob {

	return &PRMergedJob{
		jobId:        job.ID,
		githubClient: githubClient,
		pubsub:       pubsub,
		taskRepo:     taskRepo,
		jobRepo:      jobRepo,
		testflowRepo: testflowRepo,
		logger:       tagCreateLog.With("type", "PRMergedJob", "job", job.ID, "testflow", job.TestFlowId),
	}
}

func (prMerged *PRMergedJob) Id() string {
	return prMerged.jobId.Hex()
}

func (prMerged *PRMergedJob) RunImmediately(ctx context.Context) (primitive.ObjectID, error) {
	job, err := prMerged.jobRepo.Get(ctx, prMerged.jobId)
	if err != nil {
		return primitive.NilObjectID, err
	}

	prMerged.logger.Infof("job(%s) start to running", job.Name)

	newJob, err := prMerged.jobRepo.IncExecCount(ctx, job.ID)
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("increase job %s exec count fail %w", job.ID, err)
	}

	id, err := prMerged.taskRepo.Save(ctx, &models.Task{
		ID:         primitive.NewObjectID(),
		Name:       job.Name + "-" + strconv.Itoa(newJob.ExecCount),
		JobId:      job.ID,
		TestFlowId: job.TestFlowId,
		State:      models.Init,
		TestId:     types.TestId(uuid.New().String()[:8]),
		BaseTime:   models.BaseTime{},
	})
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("job %s save task fail %w", job.ID, err)
	}

	prMerged.logger.Infof("job %s save task %s", job.ID, id)
	return id, nil
}

func (prMerged *PRMergedJob) execTag(ctx context.Context, pushEvent *github.PullRequestEvent) error {
	job, err := prMerged.jobRepo.Get(ctx, prMerged.jobId)
	if err != nil {
		return err
	}

	fullName := pushEvent.GetRepo().GetFullName()

	var matched bool
	for _, match := range job.PRMergedEventMatchs {
		if strings.Contains(match.Repo, fullName) {
			sourceMatched, err := regexp.MatchString(match.SourcePattern, types.GetString(pushEvent.GetPullRequest().Head.Ref))
			if err != nil {
				return err
			}

			destMatched, err := regexp.MatchString(match.BasePattern, types.GetString(pushEvent.GetPullRequest().Base.Ref))
			if err != nil {
				return err
			}
			if !sourceMatched || !destMatched {
				continue
			}
			matched = true
			break
		}
	}

	if !matched {
		return nil
	}

	_, err = prMerged.generateTaskFromJob(ctx, job)
	return err
}

func (prMerged *PRMergedJob) Run(ctx context.Context) error {
	go func() {
		for event := range prMerged.pubsub.Sub(modules.PR_MERGED_TOPIC) {
			err := prMerged.execTag(ctx, event.(*github.PullRequestEvent))
			if err != nil {
				prMerged.logger.Infof("tag exec fail %v", err)
			}
		}
	}()
	return nil
}

func (prMerged *PRMergedJob) generateTaskFromJob(ctx context.Context, job *models.Job) (primitive.ObjectID, error) {
	newJob, err := prMerged.jobRepo.IncExecCount(ctx, prMerged.jobId)
	if err != nil {
		return primitive.NilObjectID, err
	}

	// all use master branch, do not inherit version for job
	id, err := prMerged.taskRepo.Save(ctx, &models.Task{
		ID:         primitive.NewObjectID(),
		Name:       job.Name + "-" + strconv.Itoa(newJob.ExecCount),
		JobId:      job.ID,
		TestFlowId: job.TestFlowId,
		State:      models.Init,
		TestId:     types.TestId(uuid.New().String()[:8]),
		BaseTime:   models.BaseTime{},
	})
	if err != nil {
		return primitive.NilObjectID, err
	}
	prMerged.logger.Infof("save tag %s", id)
	return id, nil
}

func (prMerged *PRMergedJob) Stop(_ context.Context) error {

	return nil
}

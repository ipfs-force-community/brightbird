package job

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v51/github"
	"github.com/google/uuid"
	logging "github.com/ipfs/go-log/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	"github.com/ipfs-force-community/brightbird/internal/backend/modules"
	"github.com/ipfs-force-community/brightbird/pkg/models"
	"github.com/ipfs-force-community/brightbird/pkg/repo"
	"github.com/ipfs-force-community/brightbird/pkg/types"
)

var prMergedjobLog = logging.Logger("tag_create_job")

var _ IJob = (*PRMergedJob)(nil)

type PRMergedJob struct {
	jobId primitive.ObjectID

	taskRepo     repo.ITaskRepo
	jobRepo      repo.IJobRepo
	testflowRepo repo.ITestFlowRepo
	pluginRepo   repo.IPluginService

	githubClient *github.Client
	pubsub       modules.WebHookPubsub

	logger *zap.SugaredLogger
}

func NewPRMergedJob(job models.Job, pluginRepo repo.IPluginService, pubsub modules.WebHookPubsub, githubClient *github.Client, taskRepo repo.ITaskRepo, jobRepo repo.IJobRepo, testflowRepo repo.ITestFlowRepo) *PRMergedJob {

	return &PRMergedJob{
		jobId:        job.ID,
		githubClient: githubClient,
		pluginRepo:   pluginRepo,
		pubsub:       pubsub,
		taskRepo:     taskRepo,
		jobRepo:      jobRepo,
		testflowRepo: testflowRepo,
		logger:       prMergedjobLog.With("type", "PRMergedJob", "job", job.ID, "testflow", job.TestFlowId),
	}
}

func (prMerged *PRMergedJob) ID() string {
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
	testflow, err := prMerged.testflowRepo.Get(ctx, &repo.GetTestFlowParams{
		ID: job.TestFlowId,
	})
	if err != nil {
		return err
	}
	graph := &models.Graph{}
	err = yaml.Unmarshal([]byte(testflow.Graph), graph)
	if err != nil {
		return err
	}

	fullName := pushEvent.GetRepo().GetFullName()

	var matched bool
	for _, match := range job.PRMergedEventMatches {
		if !strings.Contains(match.Repo, fullName) {
			continue
		}
		sourceMatched, err := regexp.MatchString(match.SourcePattern, types.GetString(pushEvent.GetPullRequest().Head.Ref))
		if err != nil {
			return err
		}
		destMatched, err := regexp.MatchString(match.BasePattern, types.GetString(pushEvent.GetPullRequest().Base.Ref))
		if err != nil {
			return err
		}
		if !sourceMatched || !destMatched || pushEvent.GetPullRequest().MergeCommitSHA == nil {
			continue
		}

		mergeCommitSha := *pushEvent.GetPullRequest().MergeCommitSHA
		for _, node := range graph.Pipeline {
			if node.Value.PluginType != types.Deploy {
				continue
			}
			plugin, err := prMerged.pluginRepo.GetPlugin(ctx, node.Value.Name, node.Value.Version)
			if err != nil {
				return err
			}
			if match.Repo == plugin.Repo {
				job.Versions[plugin.Name] = mergeCommitSha
				matched = true
			}
		}
	}

	if !matched {
		prMerged.logger.Warnf("not get matched for pr_merge, do nothing")
		return nil
	}

	// save job changed
	if _, err = prMerged.jobRepo.Save(ctx, job); err != nil {
		return err
	}
	// generate task
	_, err = prMerged.generateTaskFromJob(ctx, job)
	return err
}

func (prMerged *PRMergedJob) Run(ctx context.Context) error {
	go func() {
		for event := range prMerged.pubsub.Sub(modules.PRMergeTopic) {
			err := prMerged.execTag(ctx, event.(*github.PullRequestEvent))
			if err != nil {
				prMerged.logger.Infof("tag exec fail %v", err)
			}
		}
	}()
	return nil
}

func (prMerged *PRMergedJob) NextNSchedule(_ context.Context, _ int) ([]time.Time, error) {
	return []time.Time{}, nil
}

func (prMerged *PRMergedJob) generateTaskFromJob(ctx context.Context, job *models.Job) (primitive.ObjectID, error) {
	newJob, err := prMerged.jobRepo.IncExecCount(ctx, prMerged.jobId)
	if err != nil {
		return primitive.NilObjectID, err
	}

	// all use master branch, do not inherit version for job
	id, err := prMerged.taskRepo.Save(ctx, &models.Task{
		ID:              primitive.NewObjectID(),
		Name:            job.Name + "-" + strconv.Itoa(newJob.ExecCount),
		JobId:           job.ID,
		TestFlowId:      job.TestFlowId,
		State:           models.Init,
		TestId:          types.TestId(uuid.New().String()[:8]),
		BaseTime:        models.BaseTime{},
		InheritVersions: job.Versions,
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

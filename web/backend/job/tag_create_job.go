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
	giturls "github.com/whilp/git-urls"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

var tagCreateLog = logging.Logger("tag_create_job")

var _ IJob = (*TagCreateJob)(nil)

type TagCreateJob struct {
	jobId        primitive.ObjectID
	taskRepo     repo.ITaskRepo
	jobRepo      repo.IJobRepo
	testflowRepo repo.ITestFlowRepo

	pluginRepo   repo.IPluginService
	githubClient *github.Client
	pubsub       modules.WebHookPubsub

	logger *zap.SugaredLogger
}

func NewTagCreateJob(job models.Job, pluginRepo repo.IPluginService, pubsub modules.WebHookPubsub, githubClient *github.Client, taskRepo repo.ITaskRepo, jobRepo repo.IJobRepo, testflowRepo repo.ITestFlowRepo) *TagCreateJob {

	return &TagCreateJob{
		jobId:        job.ID,
		githubClient: githubClient,
		pubsub:       pubsub,
		taskRepo:     taskRepo,
		jobRepo:      jobRepo,
		testflowRepo: testflowRepo,
		pluginRepo:   pluginRepo,
		logger:       tagCreateLog.With("type", "TagCreatedJob", "job", job.ID, "testflow", job.TestFlowId),
	}
}

func (tagCreateJob *TagCreateJob) Id() string {
	return tagCreateJob.jobId.Hex()
}

func (tagCreateJob *TagCreateJob) RunImmediately(ctx context.Context) (primitive.ObjectID, error) {
	//get latest version and match. run with matched version
	job, err := tagCreateJob.jobRepo.Get(ctx, tagCreateJob.jobId)
	if err != nil {
		return primitive.NilObjectID, err
	}

	testflow, err := tagCreateJob.testflowRepo.Get(ctx, &repo.GetTestFlowParams{
		ID: job.TestFlowId,
	})
	if err != nil {
		return primitive.NilObjectID, err
	}

	for _, match := range job.TagCreateEventMatchs {
		owner, repoName, err := toGitOwnerAndRepo(match.Repo)
		if err != nil {
			return primitive.NilObjectID, err
		}
		tags, _, err := tagCreateJob.githubClient.Repositories.ListTags(ctx, owner, repoName, &github.ListOptions{PerPage: 50})
		if err != nil {
			return primitive.NilObjectID, err
		}

		for _, tag := range tags {
			matched, err := regexp.Match(match.TagPattern, []byte(tag.GetName()))
			if err != nil {
				return primitive.NilObjectID, err
			}
			if !matched {
				continue
			}

			for _, node := range testflow.Nodes {
				plugin, err := tagCreateJob.pluginRepo.GetPlugin(ctx, node.Name, node.Version)
				if err != nil {
					return primitive.NilObjectID, err
				}

				if match.Repo == plugin.Repo {
					job.Versions[plugin.Name] = tag.GetName()
				}
			}
			break
		}
	}
	_, err = tagCreateJob.jobRepo.Save(ctx, job)
	if err != nil {
		return primitive.NilObjectID, err
	}

	return tagCreateJob.generateTaskFromJob(ctx, job)
}

func (tagCreateJob *TagCreateJob) execTag(ctx context.Context, pushEvent *github.PushEvent) error {
	job, err := tagCreateJob.jobRepo.Get(ctx, tagCreateJob.jobId)
	if err != nil {
		return err
	}

	testflow, err := tagCreateJob.testflowRepo.Get(ctx, &repo.GetTestFlowParams{
		ID: job.TestFlowId,
	})
	if err != nil {
		return err
	}

	fullName := pushEvent.GetRepo().GetFullName()
	ref := pushEvent.GetRef()

	for _, match := range job.TagCreateEventMatchs {
		if strings.Contains(match.Repo, fullName) {
			matched, err := regexp.MatchString(match.TagPattern, ref)
			if err != nil {
				return err
			}
			if !matched {
				continue
			}
			//hit event
			//remember last hint
			for _, node := range testflow.Nodes {
				plugin, err := tagCreateJob.pluginRepo.GetPlugin(ctx, node.Name, node.Version)
				if err != nil {
					return err
				}

				if match.Repo == plugin.Repo {
					job.Versions[plugin.Name] = ref
				}
			}

			_, err = tagCreateJob.jobRepo.Save(ctx, job)
			if err != nil {
				return err
			}
			break
		}
	}

	_, err = tagCreateJob.generateTaskFromJob(ctx, job)
	return err
}

func (tagCreateJob *TagCreateJob) Run(ctx context.Context) error {
	go func() {
		for event := range tagCreateJob.pubsub.Sub(modules.CREATE_TAG_TOPIC) {
			err := tagCreateJob.execTag(ctx, event.(*github.PushEvent))
			if err != nil {
				tagCreateJob.logger.Infof("tag exec fail %v", err)
			}
		}
	}()
	return nil
}

func (tagCreateJob *TagCreateJob) generateTaskFromJob(ctx context.Context, job *models.Job) (primitive.ObjectID, error) {
	newJob, err := tagCreateJob.jobRepo.IncExecCount(ctx, tagCreateJob.jobId)
	if err != nil {
		return primitive.NilObjectID, err
	}

	id, err := tagCreateJob.taskRepo.Save(ctx, &models.Task{
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
	tagCreateJob.logger.Infof("save tag %s", id)
	return id, nil
}

func (tagCreateJob *TagCreateJob) Stop(_ context.Context) error {

	return nil
}

func toGitOwnerAndRepo(repoUrl string) (string, string, error) {
	schema, err := giturls.Parse(repoUrl)
	if err != nil {
		return "", "", err
	}

	seq := strings.Split(schema.Path[1:], "/")
	if len(seq) != 2 {
		return "", "", fmt.Errorf("uncorrect repo format %s", repoUrl)
	}
	return seq[0], seq[1], nil

}

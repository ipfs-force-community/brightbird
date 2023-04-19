package job

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

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

const githubUrl = ""

type TagCreateJob struct {
	jobId        primitive.ObjectID
	taskRepo     repo.ITaskRepo
	jobRepo      repo.IJobRepo
	testflowRepo repo.ITestFlowRepo

	deployStore  repo.DeployPluginStore
	githubClient *github.Client
	pubsub       modules.WebHookPubsub

	logger *zap.SugaredLogger
}

func NewTagCreateJob(job types.Job, deployStore repo.DeployPluginStore, pubsub modules.WebHookPubsub, githubClient *github.Client, taskRepo repo.ITaskRepo, jobRepo repo.IJobRepo) *TagCreateJob {

	return &TagCreateJob{
		jobId:        job.ID,
		githubClient: githubClient,
		pubsub:       pubsub,
		taskRepo:     taskRepo,
		jobRepo:      jobRepo,
		deployStore:  deployStore,
		logger:       tagCreateLog.With("job", job.ID, "testflow", job.TestFlowId),
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
		for _, tag := range tags {
			matched, err := regexp.Match(match.TagPattern, []byte(tag.GetName()))
			if err != nil {
				return primitive.NilObjectID, err
			}
			if !matched {
				continue
			}

			for _, node := range testflow.Nodes {
				plugin, err := tagCreateJob.deployStore.GetPlugin(node.Name)
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

func (tagCreateJob *TagCreateJob) execTag(ctx context.Context, createEvent *github.CreateEvent) error {
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

	fullName := createEvent.GetRepo().GetFullName()
	ref := createEvent.GetRef()

	for _, match := range job.TagCreateEventMatchs {
		if match.Repo == fullName {
			matched, err := regexp.Match(match.TagPattern, []byte(ref))
			if err != nil {
				return err
			}
			if !matched {
				continue
			}
			//hit event
			//remember last hint
			for _, node := range testflow.Nodes {
				plugin, err := tagCreateJob.deployStore.GetPlugin(node.Name)
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
		for event := range tagCreateJob.pubsub.Sub(modules.WEB_HOOK_TOPIC) {
			githubEvent := event.(*github.Event)
			if githubEvent.Type != nil && *githubEvent.Type == "CreateEvent" {
				payloadEvent, err := githubEvent.ParsePayload()
				if err != nil {
					tagCreateJob.logger.Errorf("parser push event failed %v", err)
					continue
				}

				createEvent := payloadEvent.(*github.CreateEvent)
				if createEvent.GetRefType() == "tag" {
					err = tagCreateJob.execTag(ctx, createEvent)
					if err != nil {
						tagCreateJob.logger.Infof("tag exec fail %v", err)
					}
				}
			}
		}
	}()
	return nil
}

func (tagCreateJob *TagCreateJob) generateTaskFromJob(ctx context.Context, job *types.Job) (primitive.ObjectID, error) {
	newJob, err := tagCreateJob.jobRepo.IncExecCount(ctx, tagCreateJob.jobId)
	if err != nil {
		return primitive.NilObjectID, err
	}

	id, err := tagCreateJob.taskRepo.Save(ctx, &types.Task{
		ID:              primitive.NewObjectID(),
		Name:            job.Name + "-" + strconv.Itoa(newJob.ExecCount),
		JobId:           job.ID,
		TestFlowId:      job.TestFlowId,
		State:           types.Init,
		TestId:          types.TestId(uuid.New().String()[:8]),
		BaseTime:        types.BaseTime{},
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

func getReposInJob(deployStore types.PluginStore, testflow types.TestFlow) ([]string, error) {
	var repos []string
	for _, node := range testflow.Nodes {
		plugin, err := deployStore.GetPlugin(node.Name)
		if err != nil {
			return nil, err
		}

		shortRepo, err := toShortRepoName(plugin.Repo)
		if err != nil {
			return nil, err
		}

		repos = append(repos, shortRepo)
	}
	return repos, nil
}

func toShortRepoName(repoUrl string) (string, error) {
	schema, err := giturls.Parse(repoUrl)
	if err != nil {
		return "", err
	}

	return schema.Path[1:], nil
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

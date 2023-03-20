package job

import (
	"context"
	"fmt"
	"github.com/bitfield/script"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/google/go-github/v50/github"
	"github.com/hunjixin/brightbird/repo"
	"github.com/hunjixin/brightbird/types"
	logging "github.com/ipfs/go-log/v2"
	giturls "github.com/whilp/git-urls"
	"os"
	"path"
	"strings"
)

var log = logging.Logger("builder")

type BuildTask struct {
	Name   string
	Repo   string
	Commit string
	Result chan error //buffer 1 length
}

type ImageBuilderMgr struct {
	proxy      string
	store      repo.DeployPluginStore
	buildSpace string
	taskCh     chan BuildTask
	//todo make build factory if need more build types
	jobMap map[string]IIMageBuilder
}

func NewImageBuilderMgr(store repo.DeployPluginStore, buildSpace, proxy string) *ImageBuilderMgr {
	return &ImageBuilderMgr{
		store:      store,
		buildSpace: buildSpace,
		proxy:      proxy,
		jobMap:     map[string]IIMageBuilder{},
		taskCh:     make(chan BuildTask),
	}
}

func (mgr *ImageBuilderMgr) BuildTestFlowEnv(ctx context.Context, deployNodes []*types.DeployNode) (map[string]string, error) {
	versionMap := make(map[string]string)
	for _, node := range deployNodes {
		plugin, err := mgr.store.GetPlugin(node.Name)
		if err != nil {
			return nil, err
		}

		codeVersionProp, err := findPropertyByName(node.Properties, types.CodeVersion)
		if err != nil {
			return nil, err
		}

		version := codeVersionProp.Value.(string)
		if version != "" {
			//get master replace back
			//todo fetch master commit id and neve save this version to database
			masterCommit, err := mgr.fetchMasterCommit(context.Background(), plugin.Repo)
			if err != nil {
				return nil, err
			}
			version = masterCommit
		}
		result := make(chan error, 1)
		mgr.taskCh <- BuildTask{
			Name:   node.Name,
			Repo:   plugin.Repo,
			Commit: version,
			Result: result,
		}
		err = <-result
		if err != nil {
			return nil, err
		}
		versionMap[node.Name] = codeVersionProp.Value.(string)
	}
	return versionMap, nil
}

func (mgr *ImageBuilderMgr) AddBuildTask(task BuildTask) {
	mgr.taskCh <- task
}

func (mgr *ImageBuilderMgr) Start(ctx context.Context) error {
	for {
		select {
		case buildTask := <-mgr.taskCh:
			builder, ok := mgr.jobMap[buildTask.Name]
			if !ok {
				log.Error("target %s not found", buildTask.Name)
				builder = &VenusImageBuilder{
					proxy: mgr.proxy,
				}
				err := builder.InitRepo(ctx, buildTask.Repo)
				if err != nil {
					log.Error("init builder for %s failed %v", buildTask.Name, err)
				}
				mgr.jobMap[buildTask.Name] = builder
				continue
			}

			err := builder.Build(ctx, buildTask.Commit)
			if err != nil {
				log.Error("build task (%s) commit (%s) %v", buildTask.Name, buildTask.Commit, err)
			}
			buildTask.Result <- err
		}
	}
}

func (mgr *ImageBuilderMgr) fetchMasterCommit(ctx context.Context, repoUrl string) (string, error) {
	schema, err := giturls.Parse(repoUrl)
	if err != nil {
		return "", err
	}

	phase := strings.Split(strings.TrimRight(schema.Path, ".git"), "/")
	project, repoName := phase[0], phase[1]
	client := github.NewClient(nil)

	repo, _, err := client.Repositories.Get(ctx, project, repoName)
	if err != nil {
		return "", err
	}

	defaultBranch := "main" //current github action
	if repo.DefaultBranch != nil {
		defaultBranch = *repo.DefaultBranch
	}

	branch, _, err := client.Repositories.GetBranch(ctx, project, repoName, defaultBranch, true)
	if err != nil {
		return "", err
	}
	commit := branch.GetCommit()
	if commit == nil {
		return "", fmt.Errorf("%s unable to find lastet commit, unknown reason", repoUrl)
	}
	return *commit.SHA, nil
}

type IIMageBuilder interface {
	InitRepo(ctx context.Context, repo string) error
	Build(ctx context.Context, commit string) error
}

type VenusImageBuilder struct {
	proxy    string
	repoPath string
	gitUrl   string
}

// InitRepo do something once for cache
func (builder *VenusImageBuilder) InitRepo(ctx context.Context, codeSpace string) error {
	_, err := git.PlainCloneContext(ctx, codeSpace, false, &git.CloneOptions{
		URL:             builder.gitUrl,
		Progress:        os.Stdout,
		InsecureSkipTLS: false,
		Depth:           1,
	})

	repoName, err := getRepoNameFromUrl(builder.gitUrl)
	builder.repoPath = path.Join(codeSpace, repoName)
	return err
}

func getRepoNameFromUrl(repoUrl string) (string, error) {
	schema, err := giturls.Parse(repoUrl)
	if err != nil {
		return "", err
	}
	phase := strings.Split(strings.TrimRight(schema.Path, ".git"), "/")
	return phase[1], nil
}

// Build exec build image once
func (builder *VenusImageBuilder) Build(ctx context.Context, commit string) error {
	repo, err := git.PlainOpen(builder.repoPath)
	if err != nil {
		return err
	}

	workTree, err := repo.Worktree()
	err = workTree.Clean(&git.CleanOptions{})
	if err != nil {
		return err
	}

	hasher := plumbing.NewHash(commit)
	err = workTree.Checkout(&git.CheckoutOptions{
		Hash: hasher,
	})
	if err != nil {
		return err
	}

	err = script.Exec(fmt.Sprintf("cd %s", builder.repoPath)).
		Exec(fmt.Sprintf("make docker-push TAG=%s BUILD_DOCKER_PROXY=%s", commit, builder.proxy)).Error()
	if err != nil {
		return err
	}
	return nil
}

func findPropertyByName(properties []*types.Property, name string) (*types.Property, error) {
	for _, p := range properties {
		if p.Name == name {
			return p, nil
		}
	}
	return nil, fmt.Errorf("property %s not found", name)
}

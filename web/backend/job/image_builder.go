package job

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"

	"github.com/ipfs-force-community/brightbird/models"
	"github.com/ipfs-force-community/brightbird/repo"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/web/backend/config"
	logging "github.com/ipfs/go-log/v2"
	giturls "github.com/whilp/git-urls"
	"golang.org/x/sync/errgroup"
)

var log = logging.Logger("builder")

type BuildResult struct {
	Version string
	Err     error
}
type BuildTask struct {
	Name    string
	Version string //Note plugin version
	Repo    string
	Commit  string
	Result  chan BuildResult //buffer 1 length
}

// ////////////// Builder Worker Provider  ////////////////
type IBuilderWorkerProvider interface {
	CreateBuildWorker(ctx context.Context, logger logging.StandardLogger, cfg config.BuildWorkerConfig) (IBuilderWorker, error)
}

type BuildWorkerProvider struct {
	proxy           string
	dockerOp        IDockerOperation
	ffi             FFIDownloader
	privateRegistry types.PrivateRegistry
	pluginRepo      repo.IPluginService
}

func NewBuildWorkerProvider(dockerOp IDockerOperation, pluginRepo repo.IPluginService, ffi FFIDownloader, privateRegistry types.PrivateRegistry, proxy string) *BuildWorkerProvider {
	return &BuildWorkerProvider{
		pluginRepo:      pluginRepo,
		dockerOp:        dockerOp,
		proxy:           proxy,
		ffi:             ffi,
		privateRegistry: privateRegistry,
	}
}

func (provider *BuildWorkerProvider) CreateBuildWorker(ctx context.Context, logger logging.StandardLogger, builderCfg config.BuildWorkerConfig) (IBuilderWorker, error) {
	return &BuildWorker{
		dockerOp:        provider.dockerOp,
		cfg:             builderCfg,
		proxy:           provider.proxy,
		buildMap:        map[string]IIMageBuilder{},
		ffi:             provider.ffi,
		privateRegistry: provider.privateRegistry,
		logger:          logger,
		pluginRepo:      provider.pluginRepo,
	}, nil
}

// ////////////// Builder Manager  ////////////////
type ImageBuilderMgr struct {
	pluginRepo repo.IPluginService
	workerCfgs []config.BuildWorkerConfig
	taskCh     chan *BuildTask
	provider   IBuilderWorkerProvider
}

func NewImageBuilderMgr(pluginRepo repo.IPluginService, provider IBuilderWorkerProvider, workerCfgs []config.BuildWorkerConfig) *ImageBuilderMgr {
	return &ImageBuilderMgr{
		pluginRepo: pluginRepo,
		workerCfgs: workerCfgs,
		provider:   provider,
		taskCh:     make(chan *BuildTask, len(workerCfgs)*2),
	}
}

func (mgr *ImageBuilderMgr) BuildTestFlowEnv(ctx context.Context, deployNodes []models.PipelineItem, versions map[string]string) (map[string]string, error) {
	versionMap := make(map[string]string)
	mapLk := sync.Mutex{}

	g, _ := errgroup.WithContext(ctx)
	for _, node := range deployNodes {
		if node.Value.PluginType != types.Deploy {
			continue
		}
		nodeCpy := *node.Value //copy
		g.Go(func() error {
			plugin, err := mgr.pluginRepo.GetPlugin(ctx, nodeCpy.Name, nodeCpy.Version)
			if err != nil {
				return err
			}

			result := make(chan BuildResult, 1)
			mgr.taskCh <- &BuildTask{
				Name:    nodeCpy.Name,
				Version: nodeCpy.Version,
				Repo:    plugin.Repo,
				Commit:  versions[nodeCpy.Name],
				Result:  result,
			}
			br := <-result
			if br.Err != nil {
				return fmt.Errorf("build %s failed reason: %v", nodeCpy.Name, br.Err)
			}
			mapLk.Lock()
			defer mapLk.Unlock()
			versionMap[nodeCpy.Name] = br.Version
			return nil
		})
	}

	err := g.Wait()
	if err != nil {
		return nil, err
	}

	return versionMap, nil
}

func (mgr *ImageBuilderMgr) AddBuildTask(task *BuildTask) {
	mgr.taskCh <- task
}

func (mgr *ImageBuilderMgr) Start(ctx context.Context) error {
	//run worker
	for _, workerCfg := range mgr.workerCfgs {
		builder, err := mgr.provider.CreateBuildWorker(ctx, log.With("worker", workerCfg.BuildSpace), workerCfg)
		if err != nil {
			return err
		}
		go builder.Start(ctx, mgr.taskCh)
	}
	return nil
}

// ////////////// Builder Worker  ////////////////

type IBuilderWorker interface {
	Start(ctx context.Context, taskCh <-chan *BuildTask)
}

// BuildWorker implement for local build
type BuildWorker struct {
	proxy           string
	pluginRepo      repo.IPluginService
	dockerOp        IDockerOperation
	buildMap        map[string]IIMageBuilder
	ffi             FFIDownloader
	privateRegistry types.PrivateRegistry

	cfg    config.BuildWorkerConfig
	logger logging.StandardLogger
}

func (worker *BuildWorker) Start(ctx context.Context, taskCh <-chan *BuildTask) {
	worker.logger.Infof("worker start wait build task")
	for buildTask := range taskCh {
		version, err := worker.do(ctx, buildTask)
		buildTask.Result <- BuildResult{
			Version: version,
			Err:     err,
		}
	}
}

func (worker *BuildWorker) do(ctx context.Context, buildTask *BuildTask) (string, error) {
	worker.logger.Infof("receive task %s commit %s", buildTask.Name, buildTask.Commit)
	plugin, err := worker.pluginRepo.GetPlugin(ctx, buildTask.Name, buildTask.Version)
	if err != nil {
		return "", err
	}

	builder, ok := worker.buildMap[plugin.Repo]
	if !ok {
		worker.logger.Infof("target %s not found and create a new builder", buildTask.Name)
		builder = &VenusImageBuilder{
			proxy:           worker.proxy,
			codeSpace:       worker.cfg.BuildSpace,
			ffi:             worker.ffi,
			privateRegistry: string(worker.privateRegistry),
		}
		err := builder.InitRepo(context.Background(), plugin.Repo)
		if err != nil {
			worker.logger.Errorf("init builder for %s failed %v", buildTask.Name, err)
			return "", err
		}
		worker.buildMap[plugin.Repo] = builder
	}

	//get master replace back
	//todo fetch master commit id and neve save this version to database
	version, err := builder.FetchCommit(ctx, buildTask.Commit)
	if err != nil {
		return "", err
	}
	worker.logger.Infof("get repo %s commit %s success", buildTask.Name, version)

	//check if images exit
	hasImage, err := worker.dockerOp.CheckImageExit(ctx, fmt.Sprintf("filvenus/%s", plugin.ImageTarget), version)
	if err != nil {
		return "", fmt.Errorf("check image %w", err)
	}

	if !hasImage {
		worker.logger.Infof("try to build repo %s commit %s success", buildTask.Name, plugin.ImageTarget, version)
		err := builder.Build(ctx, version)
		if err != nil {
			worker.logger.Errorf("build task (%s) commit (%s) %v", buildTask.Name, version, err)
			return "", err
		}
	} else {
		worker.logger.Debugf("node %s (%s) have build image before skip", buildTask.Name, version)
	}

	return version, nil
}

type IIMageBuilder interface {
	InitRepo(ctx context.Context, repo string) error
	FetchCommit(ctx context.Context, commit string) (string, error)
	Build(ctx context.Context, commit string) error
}

type VenusImageBuilder struct {
	proxy     string
	codeSpace string
	repoPath  string

	repo *git.Repository
	ffi  FFIDownloader

	privateRegistry string
}

// InitRepo do something once for cache
func (builder *VenusImageBuilder) InitRepo(ctx context.Context, repoUrl string) error {
	repoName, err := getRepoNameFromUrl(repoUrl)
	if err != nil {
		return err
	}

	builder.repoPath = path.Join(builder.codeSpace, repoName)
	_, err = os.Stat(builder.repoPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	builder.repo, err = git.PlainOpen(builder.repoPath)
	if err == nil {
		return nil
	}

	log.Warnf("open git repo %s fail, clean and clone (%s) again %v", repoUrl, builder.repoPath, err)
	err = os.RemoveAll(builder.repoPath)
	if err != nil {
		return err
	}

	sshFormat, err := toSSHFormat(repoUrl)
	if err != nil {
		return err
	}

	_, err = git.PlainCloneContext(ctx, builder.repoPath, false, &git.CloneOptions{
		URL:             sshFormat,
		Progress:        os.Stdout,
		InsecureSkipTLS: false,
		//	Depth:           200, //should be enough
	})
	if err != nil && err != git.ErrRepositoryAlreadyExists {
		_ = os.RemoveAll(builder.repoPath) //clean fail repo
		return err
	}

	//check again
	builder.repo, err = git.PlainOpen(builder.repoPath)
	return err
}

func (builder *VenusImageBuilder) updateRepo(ctx context.Context) error {
	workTree, err := builder.repo.Worktree()
	if err != nil {
		return err
	}

	err = workTree.Clean(&git.CleanOptions{})
	if err != nil && err != plumbing.ErrObjectNotFound {
		return err
	}

	err = builder.repo.FetchContext(ctx, &git.FetchOptions{
		Progress:        os.Stdout,
		InsecureSkipTLS: true,
		Force:           true,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return err
	}

	//exec git pull on main branch avoid confict on specific branch
	branches, err := builder.repo.Branches()
	if err != nil {
		return err
	}

	masterBranch := "master"
	err = branches.ForEach(func(branch *plumbing.Reference) error {
		switch true {
		case branch.Name().Short() == "master":
			masterBranch = "master"
		case branch.Name().Short() == "trunk":
			masterBranch = "trunk"
		default:
			masterBranch = "main"
		}
		return nil
	})
	if err != nil {
		return err
	}

	err = workTree.Checkout(&git.CheckoutOptions{Force: true, Branch: plumbing.NewBranchReferenceName(masterBranch)}) //git checkout master
	if err != nil {
		return err
	}

	log.Debugf("update repo %s branch(%s) to latest", builder.repoPath, masterBranch)
	err = workTree.PullContext(ctx, &git.PullOptions{
		Progress:        os.Stdout,
		InsecureSkipTLS: true,
		Force:           true,
	})
	if err != nil && !(err == git.ErrNonFastForwardUpdate || err == git.NoErrAlreadyUpToDate) {
		return err
	}

	builder.repo, _, err = updateSubmoduleByCmd(ctx, builder.repoPath)
	if err != nil {
		return fmt.Errorf("update submoudle fail %w", err)
	}

	return nil
}

func (builder *VenusImageBuilder) FetchCommit(ctx context.Context, commit string) (string, error) {
	err := builder.updateRepo(ctx)
	if err != nil {
		return "", err
	}

	repo, err := git.PlainOpen(builder.repoPath)
	if err != nil {
		return "", err
	}

	if len(commit) == 0 {
		//use head directly
		masterHead, err := repo.Head()
		if err != nil {
			return "", fmt.Errorf("use repo head %w", err)
		}
		return masterHead.Hash().String(), nil
	}

	//resolve commit
	hash, err := repo.ResolveRevision(plumbing.Revision(commit))
	if err == nil {
		return hash.String(), nil
	}

	if err == plumbing.ErrReferenceNotFound {
		//resolve branch or tag
		remotes, err := repo.Remotes()
		if err != nil {
			return "", fmt.Errorf("get repo remote %w", err)
		}

		//detect remote   default Origin
		remoteName := remotes[0].Config().Name
		hash, err = repo.ResolveRevision(plumbing.Revision(fmt.Sprintf("%s/%s", remoteName, commit)))
		if err != nil {
			return "", fmt.Errorf("resolve (%s) to a hash %s", commit, err)
		}
		return hash.String(), nil
	}
	return "", err
}

// Build exec build image once
func (builder *VenusImageBuilder) Build(ctx context.Context, commit string) error {
	err := builder.updateRepo(ctx)
	if err != nil {
		return err
	}

	repo, err := git.PlainOpen(builder.repoPath)
	if err != nil {
		return err
	}

	workTree, err := repo.Worktree()
	if err != nil {
		return err
	}

	hash, err := repo.ResolveRevision(plumbing.Revision(commit))
	if err != nil {
		return fmt.Errorf("get comit %w", err)
	}

	err = workTree.Checkout(&git.CheckoutOptions{
		Hash:   *hash,
		Branch: "",
		Force:  true,
	})
	if err != nil {
		return fmt.Errorf("repo checkout %w", err)
	}

	err = execMakefile(builder.repoPath, "make", "docker-push", "TAG="+commit, "BUILD_DOCKER_PROXY="+builder.proxy, "PRIVATE_REGISTRY="+builder.privateRegistry)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	return nil
}

// have bug https://github.com/go-git/go-git/issues/511
func updateSubmoduleByCmd(ctx context.Context, dir string) (*git.Repository, *git.Worktree, error) {
	err := execMakefile(dir, "git", "submodule", "update", "--init", "--recursive")
	if err != nil {
		return nil, nil, err
	}

	repo, err := git.PlainOpen(dir)
	if err != nil {
		return nil, nil, err
	}

	workTree, err := repo.Worktree()
	if err != nil {
		return nil, nil, err
	}
	return repo, workTree, nil
}

func getRepoNameFromUrl(repoUrl string) (string, error) {
	schema, err := giturls.Parse(repoUrl)
	if err != nil {
		return "", err
	}
	phase := strings.Split(strings.TrimSuffix(schema.Path, ".git"), "/")
	return phase[2], nil
}

func toSSHFormat(repoUrl string) (string, error) {
	schema, err := giturls.Parse(repoUrl)
	if err != nil {
		return "", err
	}
	if schema.Scheme == "ssh" {
		return repoUrl, nil
	}

	return fmt.Sprintf("git@github.com:%s", schema.Path[1:]), nil
}

func execMakefile(dir string, name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.Dir = dir
	cmd.Env = os.Environ()

	//var out bytes.Buffer
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

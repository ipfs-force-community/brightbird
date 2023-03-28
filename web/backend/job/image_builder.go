package job

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/google/go-github/v50/github"
	"github.com/hunjixin/brightbird/repo"
	"github.com/hunjixin/brightbird/types"
	logging "github.com/ipfs/go-log/v2"
	giturls "github.com/whilp/git-urls"
)

var log = logging.Logger("builder")

type BuildResult struct {
	Version string
	Err     error //buffer 1 length
}
type BuildTask struct {
	Name   string
	Repo   string
	Commit string
	Result chan BuildResult //buffer 1 length
}

type ImageBuilderMgr struct {
	proxy      string
	store      repo.DeployPluginStore
	buildSpace string
	taskCh     chan BuildTask
	//todo make build factory if need more build types
	jobMap map[string]IIMageBuilder

	dockerOp IDockerOperation

	//make inject
	ffi             *ffiDownloader
	privateRegistry string
}

func NewImageBuilderMgr(dockerOp IDockerOperation, store repo.DeployPluginStore, buildSpace, proxy, githubToken, privateRegistry string) *ImageBuilderMgr {
	return &ImageBuilderMgr{
		dockerOp:        dockerOp,
		store:           store,
		buildSpace:      buildSpace,
		proxy:           proxy,
		jobMap:          map[string]IIMageBuilder{},
		taskCh:          make(chan BuildTask),
		ffi:             newFFIDownloader(githubToken),
		privateRegistry: privateRegistry,
	}
}

func (mgr *ImageBuilderMgr) BuildTestFlowEnv(ctx context.Context, deployNodes []*types.DeployNode, versions map[string]string) (map[string]string, error) {
	versionMap := make(map[string]string)
	for _, node := range deployNodes {
		plugin, err := mgr.store.GetPlugin(node.Name)
		if err != nil {
			return nil, err
		}

		result := make(chan BuildResult, 1)
		mgr.taskCh <- BuildTask{
			Name:   node.Name,
			Repo:   plugin.Repo,
			Commit: versions[node.Name],
			Result: result,
		}
		br := <-result
		if br.Err != nil {
			return nil, br.Err
		}
		versionMap[node.Name] = br.Version
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
			plugin, err := mgr.store.GetPlugin(buildTask.Name)
			if err != nil {
				buildTask.Result <- BuildResult{
					Err: err,
				}
				continue
			}

			builder, ok := mgr.jobMap[buildTask.Name]
			if !ok {
				log.Infof("target %s not found and create a new builder", buildTask.Name)
				builder = &VenusImageBuilder{
					proxy:           mgr.proxy,
					codeSpace:       mgr.buildSpace,
					ffi:             mgr.ffi,
					privateRegistry: mgr.privateRegistry,
				}
				err := builder.InitRepo(context.Background(), plugin.Repo)
				if err != nil {
					log.Errorf("init builder for %s failed %v", buildTask.Name, err)
					buildTask.Result <- BuildResult{
						Err: err,
					}
					continue
				}
				mgr.jobMap[buildTask.Name] = builder
			}

			//get master replace back
			//todo fetch master commit id and neve save this version to database
			version, err := builder.FetchCommit(ctx, buildTask.Commit)
			if err != nil {
				buildTask.Result <- BuildResult{
					Err: err,
				}
				continue
			}

			//check if images exit
			hasImage, err := mgr.dockerOp.CheckImageExit(ctx, fmt.Sprintf("filvenus/%s", plugin.ImageTarget), version)
			if err != nil {
				buildTask.Result <- BuildResult{
					Err: err,
				}
				continue
			}

			if !hasImage {
				err := builder.Build(ctx, version)
				if err != nil {
					log.Errorf("build task (%s) commit (%s) %v", buildTask.Name, version, err)
					buildTask.Result <- BuildResult{
						Err: err,
					}
					continue
				}
			}

			buildTask.Result <- BuildResult{
				Version: version,
			}
		}
	}
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
	ffi  *ffiDownloader

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
		return err
	}

	log.Errorf("open git repo %s faile, clean and clone again %v", repoUrl, err)
	err = os.RemoveAll(builder.repoPath)
	if err != nil {
		return err
	}

	_, err = git.PlainCloneContext(ctx, builder.repoPath, false, &git.CloneOptions{
		URL:             repoUrl,
		Progress:        os.Stdout,
		InsecureSkipTLS: false,
	})
	if err != nil && err != git.ErrRepositoryAlreadyExists {
		return err
	}

	//check again
	builder.repo, err = git.PlainOpen(builder.repoPath)
	if err == nil {
		return err
	}

	return nil
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

	err = builder.repo.Fetch(&git.FetchOptions{
		Progress: os.Stdout,
		// InsecureSkipTLS skips ssl verify if protocol is https
		InsecureSkipTLS: false,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return err
	}

	err = updateSubmodule(workTree)
	if err != nil {
		return err
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

	workTree, err := repo.Worktree()
	if err != nil {
		return "", err
	}

	if len(commit) == 0 {
		err = workTree.Checkout(&git.CheckoutOptions{Force: true}) //git checkout master
		if err != nil {
			return "", err
		}
		headHash, err := repo.Head()
		if err != nil {
			return "", err
		}
		return headHash.String(), nil
	}

	hash, err := repo.ResolveRevision(plumbing.Revision(commit))
	if err == nil {
		return hash.String(), nil
	}

	if err == plumbing.ErrReferenceNotFound {
		remotes, err := repo.Remotes()
		if err != nil {
			return "", err
		}
		//detact remote
		remoteName := remotes[0].Config().Name
		hash, err = repo.ResolveRevision(plumbing.Revision(fmt.Sprintf("%s/%s", remoteName, commit)))
		if err != nil {
			return "", err
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
		return err
	}

	err = workTree.Checkout(&git.CheckoutOptions{
		Hash:   *hash,
		Branch: "",
		Force:  true,
	})
	if err != nil {
		return err
	}

	err = updateSubmodule(workTree)
	if err != nil {
		return err
	}

	submodules, err := workTree.Submodules()
	if err != nil {
		return err
	}

	ffiVersion := ""
	for _, module := range submodules {
		if strings.Contains(module.Config().Name, "filecoin-ffi") {
			status, err := module.Status()
			if err != nil {
				return err
			}
			ffiVersion = status.Expected.String()
			break
		}
	}

	if len(ffiVersion) > 0 {
		ffiPath, err := builder.ffi.downloadFFI(ctx, ffiVersion)
		if err != nil {
			return err
		}

		//	tar -C "${__tmp_dir}" -xzf "${__tarball_path}"
		err = execMakefile(builder.repoPath, "tar", "-xzf", ffiPath, "-C", "./extern/filecoin-ffi")
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
	}

	err = execMakefile(builder.repoPath, "make", "docker-push", "TAG="+commit, "BUILD_DOCKER_PROXY="+builder.proxy, "PRIVATE_REGISTRY="+builder.privateRegistry)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	return nil
}

func updateSubmodule(workTree *git.Worktree) error {
	submodules, err := workTree.Submodules()
	if err != nil {
		return err
	}
	for _, sub := range submodules {
		status, err := sub.Status()
		if err != nil {
			return err
		}
		if status.IsClean() {
			err = sub.Update(&git.SubmoduleUpdateOptions{
				Init: true,
				// NoFetch tell to the update command to not fetch new objects from the
				// remote site.
				NoFetch:           true,
				RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func getRepoNameFromUrl(repoUrl string) (string, error) {
	schema, err := giturls.Parse(repoUrl)
	if err != nil {
		return "", err
	}
	phase := strings.Split(strings.TrimRight(schema.Path, ".git"), "/")
	return phase[2], nil
}

func execMakefile(dir string, name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.Dir = dir

	//var out bytes.Buffer
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

type ffiDownloader struct {
	tempPath string
	token    string
}

func newFFIDownloader(token string) *ffiDownloader {
	return &ffiDownloader{
		tempPath: os.TempDir(),
		token:    token,
	}
}
func (downloader ffiDownloader) downloadFFI(ctx context.Context, releaseTag string) (string, error) {
	fileName := releaseTag + "-filecoin-ffi-Linux-standard.tar.gz"
	filePath := path.Join(downloader.tempPath, fileName)

	_, err := os.Stat(filePath)
	if err == nil {
		return filePath, nil
	}
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}

	client := github.NewTokenClient(ctx, downloader.token)
	tag, _, err := client.Repositories.GetReleaseByTag(ctx, "filecoin-project", "filecoin-ffi", releaseTag[0:16])
	if err != nil {
		return "", err
	}

	var linuxAssert *github.ReleaseAsset
	for _, assert := range tag.Assets {
		if assert.Name != nil && strings.Contains(*assert.Name, "Linux") && assert.URL != nil {
			linuxAssert = assert
		}
	}
	if linuxAssert == nil {
		return "", fmt.Errorf("linux release for tag %s not exit", releaseTag)
	}

	body, _, err := client.Repositories.DownloadReleaseAsset(ctx, "filecoin-project", "filecoin-ffi", *linuxAssert.ID, client.Client())
	if err != nil {
		return "", err
	}

	fs, err := os.Create(filePath)
	if err != nil {
		return "", err
	}

	_, err = io.Copy(fs, body)
	if err != nil {
		return "", err
	}
	return filePath, nil
}

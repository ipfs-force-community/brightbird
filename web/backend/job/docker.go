package job

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/heroku/docker-registry-client/registry"
	"github.com/hunjixin/brightbird/web/backend/config"
	logging "github.com/ipfs/go-log/v2"
	"github.com/mittwald/goharbor-client/v5/apiv2"
	cfg2 "github.com/mittwald/goharbor-client/v5/apiv2/pkg/config"
)

func init() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
}

var dockerLog = logging.Logger("docker_hub")

type IDockerOperation interface {
	CheckImageExit(ctx context.Context, name string, tag string) (bool, error)
}

type DockerRegistry struct {
	registries []IDockerRegistry
}

var _ IDockerOperation = (*DockerRegistry)(nil)

func NewDockerRegistry(registriesCfg []config.DockerRegistry) (*DockerRegistry, error) {
	var registries []IDockerRegistry
	for _, regCfg := range registriesCfg {
		switch regCfg.Type {
		case "offical":
			hub, err := NewOfficialClientWrapper(regCfg.URL, regCfg.UserName, regCfg.Password)
			if err != nil {
				dockerLog.Errorf("connect to official registry %s %v", regCfg.URL, err)
				continue
			}
			registries = append(registries, hub)
		case "harbor":
			hub, err := NewHarborClientWrapper(regCfg.URL, regCfg.UserName, regCfg.Password)
			if err != nil {
				dockerLog.Errorf("connect to official registry %s %v", regCfg.URL, err)
				continue
			}
			registries = append(registries, hub)
		default:
			return nil, fmt.Errorf("unsupport docker registry type")
		}

	}

	if len(registries) == 0 {
		return nil, errors.New("no available docker registry")
	}

	return &DockerRegistry{registries: registries}, nil
}

func (dReg *DockerRegistry) CheckImageExit(ctx context.Context, name string, tag string) (bool, error) {
	for _, hub := range dReg.registries {
		tags, err := hub.Tags(ctx, name)
		if err != nil {
			dockerLog.Errorf("get %s's tags  from hub %s %v", name, err)
			break
		}
		for _, itag := range tags {
			if tag == itag {
				return true, nil
			}
		}
	}
	return false, nil
}

type IDockerRegistry interface {
	Tags(context.Context, string) ([]string, error)
}

var _ IDockerRegistry = (*HarborClientWrapper)(nil)

type HarborClientWrapper struct {
	c *apiv2.RESTClient
}

func NewHarborClientWrapper(apiURL, username, password string) (*HarborClientWrapper, error) {
	harborClient, err := apiv2.NewRESTClientForHost(apiURL, username, password, &cfg2.Options{
		PageSize: 100,
		Page:     1,
	})
	if err != nil {
		return nil, err
	}

	return &HarborClientWrapper{harborClient}, nil
}

func (client *HarborClientWrapper) Tags(ctx context.Context, name string) ([]string, error) {
	seq := strings.Split(name, "/")
	artifacts, err := client.c.ListArtifacts(ctx, seq[0], seq[1])
	if err != nil {
		return nil, err
	}

	var allTags []string
	for _, artifact := range artifacts {
		tags, err := client.c.ListTags(ctx, seq[0], seq[1], artifact.Digest)
		if err != nil {
			return nil, err
		}
		for _, tag := range tags {
			allTags = append(allTags, tag.Name)
		}
	}
	return allTags, nil
}

var _ IDockerRegistry = (*OfficialClientWrapper)(nil)

type OfficialClientWrapper struct {
	client *registry.Registry
}

func NewOfficialClientWrapper(apiURL, username, password string) (*OfficialClientWrapper, error) {
	hub, err := registry.NewInsecure(apiURL, username, password)
	if err != nil {
		return nil, err
	}

	return &OfficialClientWrapper{hub}, nil
}

func (client *OfficialClientWrapper) Tags(ctx context.Context, name string) ([]string, error) {
	return client.client.Tags(name)
}

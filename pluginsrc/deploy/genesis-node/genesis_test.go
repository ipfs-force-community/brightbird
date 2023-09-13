package genesisnode

import (
	"fmt"
	"testing"

	"github.com/ipfs-force-community/brightbird/env"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	yaml_k8s "k8s.io/apimachinery/pkg/util/yaml"
)

func TestDeployFromConfig(t *testing.T) {
	configMapFs, err := f.Open("genesis/genesis-stateful-deployment.yaml")
	assert.NoError(t, err)

	data, err := env.QuickRender(configMapFs, RenderParams{
		Config: Config{
			BaseConfig: env.NewBaseConfig("aaaaa", "fffffff"),
		},
		UniqueId:  "aaaaaa",
		NameSpace: "default",
		Registry:  "192.168.1.1.",
	})
	assert.NoError(t, err)

	configMap := &corev1.ConfigMap{}
	fmt.Println(string(data))
	err = yaml_k8s.Unmarshal(data, configMap)
	assert.NoError(t, err)
}

func TestParseLibp2pUrl(t *testing.T) {
	mr, err := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/34567/p2p/12D3KooWGhZAKwWhEJr8RyWGVMtucZn7oX2ZWMVHf8TvjC8zRL7k")
	assert.NoError(t, err)

	port, err := mr.ValueForProtocol(ma.P_TCP)
	assert.NoError(t, err)

	peer, err := mr.ValueForProtocol(ma.P_P2P)
	assert.NoError(t, err)
	fmt.Println(port)
	fmt.Println(peer)
}

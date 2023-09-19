package damoclesmanager

import (
	"fmt"
	"testing"

	"github.com/filecoin-project/go-address"
	"github.com/ipfs-force-community/brightbird/env"
	"github.com/stretchr/testify/assert"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	yaml_k8s "k8s.io/apimachinery/pkg/util/yaml"
)

func TestConfigmapFromConfig(t *testing.T) {
	configMapFs, err := f.Open("damocles-manager-exist/damocles-manager-exist-configmap.yaml")
	assert.NoError(t, err)

	data, err := env.QuickRender(configMapFs, RenderParams{
		Config: Config{
			BaseConfig: env.NewBaseConfig("aaaaa", "fffffff"),
			VConfig: VConfig{

				PieceStores:   []string{"default"},
				PersistStores: []string{"default"},

				NodeUrl:      "/ip4/192.168.1.1",
				MessagerUrl:  "/ip4/192.168.1.1",
				MarketUrl:    "/ip4/192.168.1.1",
				GatewayUrl:   "/ip4/192.168.1.1",
				MinerAddress: "f0100",

				SenderWalletAddress: address.NewForTestGetter()(),
				UserToken:           "tokenabc",
			},
		},
		UniqueId:  "aaaaaa",
		NameSpace: "default",
		Registry:  "192.168.1.1.",
		Args:      []string{},
	})
	assert.NoError(t, err)

	configMap := &corev1.ConfigMap{}
	fmt.Println(string(data))
	err = yaml_k8s.Unmarshal(data, configMap)
	assert.NoError(t, err)
}

func TestDeployFromConfig(t *testing.T) {
	configMapFs, err := f.Open("damocles-manager-exist/damocles-manager-exist-statefulset.yaml")
	assert.NoError(t, err)

	data, err := env.QuickRender(configMapFs, RenderParams{
		Config: Config{
			BaseConfig: env.NewBaseConfig("aaaaa", "fffffff"),
			VConfig: VConfig{

				PieceStores:   []string{"default"},
				PersistStores: []string{"default"},

				NodeUrl:      "/ip4/192.168.1.1",
				MessagerUrl:  "/ip4/192.168.1.1",
				MarketUrl:    "/ip4/192.168.1.1",
				GatewayUrl:   "/ip4/192.168.1.1",
				MinerAddress: "f0100",

				SenderWalletAddress: address.NewForTestGetter()(),
				UserToken:           "tokenabc",
			},
		},
		UniqueId:  "aaaaaa",
		NameSpace: "default",
		Registry:  "192.168.1.1.",
		Args:      []string{},
	})
	assert.NoError(t, err)

	statefulset := &appv1.StatefulSet{}
	fmt.Println(string(data))
	err = yaml_k8s.Unmarshal(data, statefulset)
	assert.NoError(t, err)
}

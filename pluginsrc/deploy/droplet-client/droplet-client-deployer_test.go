package dropletclient

import (
	"fmt"
	"testing"

	"github.com/ipfs-force-community/brightbird/env"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	yaml_k8s "k8s.io/apimachinery/pkg/util/yaml"
)

func TestConfigmapFromConfig(t *testing.T) {
	configMapFs, err := f.Open("droplet-client/droplet-client-configmap.yaml")
	assert.NoError(t, err)

	data, err := env.QuickRender(configMapFs, RenderParams{
		Config: Config{
			BaseConfig: env.NewBaseConfig("aaaaa", "fffffff"),
			VConfig: VConfig{
				NodeUrl:              "/ip4/192.168.1.1",
				WalletUrl:            "/ip4/192.168.1.1",
				DefaultMarketAddress: "/ip4/192.168.1.1",
				UserToken:            "tokenabc",
				WalletToken:          "tokenabc",
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

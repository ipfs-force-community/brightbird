package damoclesworker

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	yaml_k8s "k8s.io/apimachinery/pkg/util/yaml"

	"github.com/ipfs-force-community/brightbird/env"
)

func TestDeployFromConfig(t *testing.T) {
	configMapFs, err := f.Open("damocles-worker/damocles-worker-configmap.yaml")
	assert.NoError(t, err)

	data, err := env.QuickRender(configMapFs, RenderParams{
		Config: Config{
			BaseConfig: env.NewBaseConfig("aaaaa", "fffffff"),
			VConfig: VConfig{

				PieceStores:        []string{"default"},
				DamoclesManagerUrl: "/ip4/192.168.1.1",
				MinerAddress:       "f0100",
				UserToken:          "tokenabc",
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

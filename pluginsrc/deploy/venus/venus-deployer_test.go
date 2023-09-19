package venus

import (
	"fmt"
	"os"
	"testing"

	"github.com/ipfs-force-community/brightbird/env"
	"github.com/stretchr/testify/assert"
	appv1 "k8s.io/api/apps/v1"
	yaml_k8s "k8s.io/apimachinery/pkg/util/yaml"
)

func TestVenusHADeployer_YAML_Check(t *testing.T) {
	f, err := os.Open("./venus-node/venus-node-stateful-deployment.yaml")
	assert.NoError(t, err)
	renderParams := RenderParams{
		UniqueId: "abc",
		Config: Config{
			VConfig: VConfig{
				Replicas:        1,
				AuthUrl:         "http://127.0.0.1:8989",
				AdminToken:      "token",
				BootstrapPeers:  []string{"/ip4/127.0.0.1/tcp/130"},
				SnapshotStorage: "shared-dir-v",
				GenesisStorage:  "shared-dir-v",
			},
		},
	}
	data, err := env.QuickRender(f, renderParams)
	assert.NoError(t, err)

	fmt.Println(string(data))
	statefulSet := &appv1.StatefulSet{}
	err = yaml_k8s.Unmarshal(data, statefulSet)
	assert.NoError(t, err)
}

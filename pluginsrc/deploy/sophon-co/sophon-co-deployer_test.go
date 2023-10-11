package sophonco

import (
	"fmt"
	"os"
	"testing"

	"github.com/ipfs-force-community/brightbird/env"
	"github.com/stretchr/testify/assert"
	appv1 "k8s.io/api/apps/v1"
	yaml_k8s "k8s.io/apimachinery/pkg/util/yaml"
)

func TestChainCoDeployer_YAML_Check(t *testing.T) {
	f, err := os.Open("./sophon-co/sophon-co-statefulset.yaml")
	assert.NoError(t, err)
	renderParams := RenderParams{
		UniqueId: "abc",
		Config: Config{
			VConfig: VConfig{
				Replicas:   1,
				AuthUrl:    "http://127.0.0.1:8989",
				AdminToken: "token",
				Nodes: []string{
					"token:/dns/pod1.example.com/tcp/80",
					"token:/dns/pod2.example.com/tcp/81"},
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

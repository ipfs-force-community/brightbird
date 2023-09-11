package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	"github.com/stretchr/testify/assert"
)

func Test_resolveInputValue(t *testing.T) {
	data := []byte(`{
		"nodes": {
		  "genesis-node-dd3898da": {
			"input": {
			  "codeVersion": "",
			  "instanceName": "genesis-node-dd3898da"
			},
			"output": {
			  "addr": "t3tehwiess4l72p5rfz6rzppx42kcp25clcxhz6mvjghhy6ulqtrom24t5tkarr443lx3e2sso6j7i7d6g6poa",
			  "bootstrapPeer": "/dns/genesis-0a657d205a1e18e5-svc/tcp/34567/perr/12D3KooWJ9jpwxH26uX58cYd3FJGDf8E78HWVsPr6Jwa4zFYVu6u",
			  "rpcUrl": "/dns/genesis-0a657d205a1e18e5-svc/tcp/1234",
			  "rpcToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJyZWFkIiwid3JpdGUiLCJzaWduIiwiYWRtaW4iXX0.7adDzLsUbf_wRdpubCMWKhaTNVrhGcoQ3PcR1RVjWa8",
			  "genesisStorage": "genesis-pvc-genesis-0a657d205a1e18e5-statefulset-0"
			}
		  },
		  "sophon-auth-b1fc9c81": {
			"input": {
			  "codeVersion": "bcfaf766433b2c745142a1de5f402280de5b1e75",
			  "instanceName": "sophon-auth-b1fc9c81",
			  "replicas": 1
			},
			"output": {
			  "mysqlDSN": "root:Aa123456@(192.168.200.175:3306)/sophon-auth-0a657d20fd22cd52?parseTime=true\u0026loc=Local\u0026charset=utf8mb4\u0026collation=utf8mb4_unicode_ci\u0026readTimeout=10s\u0026writeTimeout=10s",
			  "replicas": 1,
			  "adminToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiYWRtaW4iLCJwZXJtIjoiYWRtaW4iLCJleHQiOiIifQ.gffENLlByW9jMz0rvqg-yBPhKgCwlPipRhwYA_YnGTs",
			  "codeVersion": "bcfaf766433b2c745142a1de5f402280de5b1e75",
			  "instanceName": "sophon-auth-b1fc9c81",
			  "deployName": "sophon-auth",
			  "statefulSetName": "sophon-auth-0a657d20fd22cd52",
			  "configMapName": "sophon-auth-0a657d20fd22cd52",
			  "svcName": "sophon-auth-0a657d20fd22cd52-service",
			  "svcEndpoint": "sophon-auth-0a657d20fd22cd52-service:8989"
			}
		  }
		}
	  }`)
	ctx := &env.EnvContext{}
	err := json.Unmarshal(data, ctx)
	assert.NoError(t, err)

	type VConfig struct {
		AuthUrl        string   `jsonschema:"-" json:"authUrl"`
		AdminToken     string   `jsonschema:"-" json:"adminToken"`
		BootstrapPeers []string `jsonschema:"-" json:"bootstrapPeers"`

		NetType  string `json:"netType" jsonschema:"netType" title:"Network Type" default:"force" require:"true" description:"network type: mainnet,2k,calibrationnet,force" enum:"mainnet,2k,calibrationnet,force"`
		Replicas int    `json:"replicas"  jsonschema:"replicas" title:"Replicas" default:"1" require:"true" description:"number of replicas"`

		GenesisStorage  string `json:"genesisStorage"  jsonschema:"genesisStorage" title:"GenesisStorage" default:"" require:"true" description:"used genesis file"`
		SnapshotStorage string `json:"snapshotStorage"  jsonschema:"snapshotStorage" title:"SnapshotStorage" default:"" require:"true" description:"used to read snapshot file"`
	}

	type Config struct {
		env.BaseConfig
		VConfig
	}

	type SophonAuthDeployReturn struct { //nolint
		MysqlDSN   string `json:"mysqlDSN"`
		Replicas   int    `json:"replicas" description:"number of replicas"`
		AdminToken string `json:"adminToken"`
		env.CommonDeployParams
	}

	type DepParams struct {
		Config

		BootstrapPeers []string               `json:"bootstrapPeers" jsonschema:"bootstrapPeers" title:"BootstrapPeers" require:"true" description:"config boot peers"`
		Auth           SophonAuthDeployReturn `json:"SophonAuth" jsonschema:"SophonAuth" title:"Sophon Auth" require:"true" description:"sophon auth return"`

		GenesisStorage  string `json:"genesisStorage"  jsonschema:"genesisStorage" title:"GenesisStorage" default:"" require:"true" description:"used genesis file"`
		SnapshotStorage string `json:"snapshotStorage"  jsonschema:"snapshotStorage" title:"SnapshotStorage" default:"" require:"true" description:"used to read snapshot file"`
	}

	schema, err := plugin.ParserSchema(reflect.TypeOf(DepParams{}))
	assert.NoError(t, err)

	// 1
	// "a"
	// "{{aaaa}"
	// "[{{"xxx"}}, "x"]"
	//{"SophonAuth":"{{sophon-auth-b1fc9c81}}","bootstrapPeers":"[\"{{genesis-node-dd3898da.bootstrapPeer}}\"]","genesisStorage":"{{genesis-node-dd3898da.genesisStorage}}","netType":"force","replicas":1,"snapshotStorage":""}
	input := []byte(`{"SophonAuth":"{{sophon-auth-b1fc9c81}}","bootstrapPeers":"[\"{{genesis-node-dd3898da.bootstrapPeer}}\"]","genesisStorage":"{{genesis-node-dd3898da.genesisStorage}}","netType":"force","replicas":"1","snapshotStorage":"aaa"}`)
	ZX, err := resolveInputValue(ctx, schema, input, "aaa", "xxxxxx")
	assert.NoError(t, err)
	fmt.Println(string(ZX))
}

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/swaggest/jsonschema-go"

	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	"github.com/ipfs-force-community/brightbird/models"
	"github.com/ipfs-force-community/brightbird/repo"
	"github.com/ipfs-force-community/brightbird/types"
	jsoniter "github.com/json-iterator/go"
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v3"
)

func runGraph(ctx context.Context, cfg *Config, pluginRepo repo.IPluginService, k8sEnvParams *env.K8sInitParams, testflow *models.TestFlow, task *models.Task) error {
	graph := models.Graph{}
	err := yaml.Unmarshal([]byte(testflow.Graph), &graph)
	if err != nil {
		return err
	}

	envCtx := &env.EnvContext{
		Global: cfg.GlobalParams,
		Nodes:  make(map[string]*env.NodeContext),
	}
	for _, pip := range graph.Pipeline {
		deployPlugin, err := pluginRepo.GetPlugin(ctx, pip.Value.Name, pip.Value.Version)
		if err != nil {
			return err
		}

		var codeVersion string
		if deployPlugin.PluginType == types.Deploy {
			var ok bool
			codeVersion, ok = task.CommitMap[pip.Value.Name]
			if !ok {
				return fmt.Errorf("not found version for deploy %s", pip.Value.Name)
			}
		}

		err = runNode(k8sEnvParams, envCtx, path.Join(cfg.PluginStore, deployPlugin.Path), deployPlugin, pip.Value, codeVersion)
		if err != nil {
			plugin.RespError(err)
			return err
		}
	}
	return nil
}

func runNode(k8sEnvParams *env.K8sInitParams, envCtx *env.EnvContext, pluginPath string, pluginDef *models.PluginDef, pip *types.ExecNode, codeVersion string) error {
	currentCtx := &env.NodeContext{
		Input:  []byte("{}"),
		OutPut: []byte("{}"),
	}
	envCtx.CurrentContext = pip.InstanceName
	envCtx.Nodes[pip.InstanceName] = currentCtx

	var err error
	currentCtx.Input, err = resolveInputValue(envCtx, jsonschema.Schema(pluginDef.InputSchema), pip.Input, codeVersion, pip.InstanceName)
	if err != nil {
		return err
	}

	// standard input, standard output, and standard error.
	stdInR, stdInW, err := os.Pipe()
	if err != nil {
		return err
	}
	stdOut := bytes.NewBuffer(nil)
	stdErr := bytes.NewBuffer(nil)

	//write init params
	initParams := plugin.InitParams{
		K8sInitParams: *k8sEnvParams,
		EnvContext:    *envCtx,
	}
	initData, err := json.Marshal(initParams)
	if err != nil {
		return err
	}

	log.Debugf("invoke plugin %s params %s", pip.InstanceName, string(initData))

	plugin.RespStart(pip.InstanceName)
	cmd := exec.Command(pluginPath)
	cmd.Env = os.Environ()
	cmd.Stdin = stdInR
	cmd.Stdout = io.MultiWriter(os.Stdout, stdOut)
	cmd.Stderr = io.MultiWriter(os.Stderr, stdErr)

	_, err = stdInW.Write(initData)
	if err != nil {
		return err
	}
	_, err = stdInW.Write([]byte{'\n'})
	if err != nil {
		return err
	}

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("exec plugin %s fail err(%v) stderr(%s)", pip.InstanceName, err, stdErr.String())
	}

	stdInW.Close() //nolint

	result := plugin.GetLastJSON(stdOut.String())

	newCtx := &env.EnvContext{}
	err = json.Unmarshal([]byte(result), newCtx)
	if err != nil {
		return fmt.Errorf("plugin %s result is not json format result(%s)", pip.InstanceName, result)
	}

	plugin.RespSuccess("")
	*envCtx = *newCtx //override value
	return nil
}

func resolveInputValue(envCtx *env.EnvContext, schema jsonschema.Schema, input []byte, codeVersion, instanceName string) ([]byte, error) {
	propertyFinder := plugin.NewSchemaPropertyFinder(schema)
	var err error
	iter := jsoniter.NewIterator(jsoniter.ConfigDefault).ResetBytes(input)
	valueResolve := func() func(string, string) (interface{}, error) {
		return func(keyPath string, value string) (interface{}, error) {
			propValue := value
			if strings.HasPrefix(value, "{{") || strings.HasSuffix(value, "}}") {
				valuePath := value[2 : len(value)-2]
				depNode := valuePath

				pathSeq, err := plugin.SplitJSONPath(valuePath)
				if err != nil {
					return nil, err
				}
				if len(pathSeq) == 1 {
					node, err := envCtx.GetNode(depNode)
					if err != nil {
						return nil, err
					}
					propValue = string(node.OutPut) //do convert in front page
				} else {
					depNode = pathSeq[0].Name
					node, err := envCtx.GetNode(depNode)
					if err != nil {
						return nil, err
					}

					//support array
					valuePath = joinGjsonPath(pathSeq[1:])
					result := gjson.Get(string(node.OutPut), valuePath)
					propValue = result.Raw
					if result.Type == gjson.String {
						propValue = result.Str
					}
				}
				//get value from output value and then parser it
			}
			//convert to value
			schemaType, err := propertyFinder.FindPath(keyPath)
			if err != nil {
				return nil, err
			}
			return plugin.GetJSONValue(schemaType, propValue)
		}
	}

	iter.ResetBytes(input)
	w := bytes.NewBufferString("")
	encoder := jsoniter.NewStream(jsoniter.ConfigDefault, w, 512)
	err = IterJSON(iter, encoder, "", valueResolve())
	if err != nil {
		return nil, err
	}
	err = encoder.Flush()
	if err != nil {
		return nil, err
	}

	resultInput := make(map[string]interface{})
	err = json.Unmarshal(w.Bytes(), &resultInput)
	if err != nil {
		return nil, err
	}

	resultInput["instanceName"] = instanceName
	resultInput["codeVersion"] = codeVersion
	return json.Marshal(resultInput)
}

func joinGjsonPath(pathSeq []plugin.JSONPathSec) string {
	var strBuilder strings.Builder
	for _, path := range pathSeq {
		strBuilder.WriteRune('.')
		if path.IsIndex {
			strBuilder.WriteString(strconv.Itoa(path.Index))
		} else {
			strBuilder.WriteString(path.Name)
		}
	}
	return strings.Trim(strBuilder.String(), ".")
}

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
		plugin.RespStart(pip.Value.InstanceName)
		deployPlugin, err := pluginRepo.GetPlugin(ctx, pip.Value.Name, pip.Value.Version)
		if err != nil {
			return err
		}

		var codeVersion string
		if deployPlugin.PluginType == types.Deploy {
			if deployPlugin.Buildable() {
				var ok bool
				codeVersion, ok = task.CommitMap[pip.Value.Name]
				if !ok {
					return fmt.Errorf("not found version for deploy %s", pip.Value.Name)
				}
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

	envCtxBytes, err := json.Marshal(envCtx)
	if err != nil {
		return err
	}
	log.Debugf("env context %w", string(envCtxBytes))
	log.Debugf("input %s", string(pip.Input))

	currentCtx.Input, err = resolveInputValue(envCtx, jsonschema.Schema(pluginDef.InputSchema), pip.Input, codeVersion, pip.InstanceName)
	if err != nil {
		return fmt.Errorf("resolve %s input fail %w", pip.InstanceName, err)
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
		return fmt.Errorf("plugin %s result is not json format result(%s) %w", pip.InstanceName, result, err)
	}

	plugin.RespSuccess("")
	*envCtx = *newCtx //override value
	return nil
}

func resolveInputValue(envCtx *env.EnvContext, schema jsonschema.Schema, input []byte, codeVersion, instanceName string) ([]byte, error) {
	propertyFinder := plugin.NewSchemaPropertyFinder(schema)
	valueResolve := func() func(string, string) (interface{}, error) {
		return func(keyPath string, value string) (interface{}, error) {
			propValue := value
			depNode := value

			pathSeq := plugin.SplitJSONPath(value)
			if len(pathSeq) == 1 {
				node, err := envCtx.GetNode(depNode)
				if err != nil {
					return nil, fmt.Errorf("find node %s keyPath %s fail %w", depNode, keyPath, err)
				}
				propValue = string(node.OutPut) //do convert in front page
			} else {
				depNode = pathSeq[0].Name
				node, err := envCtx.GetNode(depNode)
				if err != nil {
					return nil, fmt.Errorf("find node %s fail keyPath %s %w", depNode, keyPath, err)
				}

				//support array
				value = joinGjsonPath(pathSeq[1:])
				result := gjson.Get(string(node.OutPut), value)
				propValue = result.Raw
				if result.Type == gjson.String { //todo
					propValue = result.Str
				}
			}

			//convert to value
			schemaType, err := propertyFinder.FindPath(keyPath)
			if err != nil {
				return nil, fmt.Errorf("resolve (%s)'s schema type fail %w", keyPath, err)
			}
			val, err := plugin.GetJSONValue(schemaType, propValue)
			if err != nil {
				return nil, fmt.Errorf("get json value (path %s, type %s) for schema %w", propValue, schemaType, err)
			}
			return val, nil
		}
	}

	// string or number and json was embed in string
	var kv map[string]interface{}
	err := json.Unmarshal(input, &kv)
	if err != nil {
		return nil, err
	}

	resultInput := make(map[string]interface{})

	for k, v := range kv {
		// 1   case 1
		// "a" case 2
		// "{{aaaa}" case 3
		// "[{{"xxx"}}, "x"]" case 4
		vStr, ok := v.(string)
		if !ok {
			// case 1   数值字面量
			resultInput[k] = v
			continue
		} else {
			schemaType, err := propertyFinder.FindPath(k)
			if err != nil {
				return nil, fmt.Errorf("resolve (%s)'s schema type fail %w", k, err)
			}
			if schemaType == jsonschema.String {
				if !strings.HasPrefix(vStr, "{{") {
					//case 2  字符串字面量
					resultInput[k] = v
					continue
				}
			}
		}
		//json类型或者变量类型
		iter := jsoniter.NewIterator(jsoniter.ConfigDefault)
		iter.ResetBytes([]byte(vStr))
		w := bytes.NewBufferString("")
		encoder := jsoniter.NewStream(jsoniter.ConfigDefault, w, 512)
		err = IterJSON(iter, encoder, k, valueResolve())
		if err != nil {
			return nil, err
		}
		err = encoder.Flush()
		if err != nil {
			return nil, err
		}

		fmt.Println(w.String())
		var val interface{}
		err = json.Unmarshal(w.Bytes(), &val)
		if err != nil {
			return nil, err
		}
		resultInput[k] = val
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

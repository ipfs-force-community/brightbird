package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	"github.com/hunjixin/brightbird/models"
	"github.com/hunjixin/brightbird/repo"
	"github.com/hunjixin/brightbird/types"
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
		Global: env.GlobalParams{
			LogLevel:         cfg.LogLevel,
			CustomProperties: cfg.CustomProperties,
		},
		Nodes: make(map[string]*env.NodeContext),
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
	iter := jsoniter.NewIterator(jsoniter.ConfigDefault).ResetBytes([]byte(currentCtx.Input))
	valueResolve := func() func(string, string) (interface{}, error) {
		return func(keyPath string, value string) (interface{}, error) {
			property, err := pluginDef.GetInputProperty(keyPath)
			if err != nil {
				return nil, err
			}
			propValue := value
			if strings.HasPrefix(string(value), "{{") || strings.HasSuffix(string(value), "}}") {
				valuePath := string(value[2 : len(value)-2])
				depNode := valuePath
				index := strings.Index(valuePath, ".")
				if index == -1 {
					node, err := envCtx.GetNode(depNode)
					if err != nil {
						return nil, err
					}
					propValue = string(node.OutPut) //do convert in front page
				} else {
					depNode = valuePath[:index]
					valuePath = valuePath[index+1:]
					node, err := envCtx.GetNode(depNode)
					if err != nil {
						return nil, err
					}

					propValue = string(gjson.Get(string(node.OutPut), valuePath).Raw)
					if err != nil {
						return nil, err
					}
				}
				//get value from output value and than parser it
			}
			//convert to value
			return plugin.GetPropertyValue(property, propValue)
		}
	}
	iter.ResetBytes(pip.Input)
	w := bytes.NewBufferString("")
	encoder := jsoniter.NewStream(jsoniter.ConfigDefault, w, 512)
	err = IterJSON(iter, encoder, "", valueResolve())
	if err != nil {
		return err
	}
	err = encoder.Flush()
	if err != nil {
		return err
	}

	resultInput := make(map[string]interface{})
	err = json.Unmarshal(w.Bytes(), &resultInput)
	if err != nil {
		return err
	}

	resultInput["instanceName"] = pip.InstanceName
	resultInput["codeVersion"] = codeVersion
	currentCtx.Input, err = json.Marshal(resultInput)
	if err != nil {
		return err
	}

	// standard input, standard output, and standard error.
	stdInR, stdInW, err := os.Pipe()
	if err != nil {
		return err
	}

	stdOutR, stdOutW, err := os.Pipe()
	if err != nil {
		return err
	}

	stdErrR, stdErrW, err := os.Pipe()
	if err != nil {
		return err
	}

	outR := bufio.NewReader(io.TeeReader(stdOutR, os.Stdout))
	readLastLine := make(chan string)
	go func() {
		var lastLine string
		for {
			thisLine, err := outR.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					readLastLine <- lastLine
					break
				}
				log.Errorf("read stdout fail %w", err)
				return
			}
			lastLine = thisLine
		}
	}()

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
	process, err := os.StartProcess(pluginPath, []string{pluginPath}, &os.ProcAttr{
		Env:   os.Environ(),
		Files: []*os.File{stdInR, stdOutW, stdErrW},
	})
	if err != nil {
		return err
	}

	_, err = stdInW.Write(initData)
	if err != nil {
		return err
	}
	_, err = stdInW.Write([]byte{'\n'})
	if err != nil {
		return err
	}

	st, err := process.Wait()
	if err != nil {
		return err
	}

	stdOutW.Close() //nolint
	stdInW.Close()  //nolint
	stdErrW.Close() //nolint

	if !st.Success() {
		r := bufio.NewReader(io.TeeReader(stdErrR, os.Stderr))
		var lastErr string
		for {
			thisLine, err := r.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}
			lastErr = thisLine
		}
		return fmt.Errorf("node exit with status %d  %s", st.ExitCode(), string(lastErr))
	}

	lastline := <-readLastLine
	newCtx := &env.EnvContext{}
	err = json.Unmarshal([]byte(lastline), newCtx)
	if err != nil {
		return err
	}

	plugin.RespSuccess("")
	*envCtx = *newCtx //override value
	return nil
}

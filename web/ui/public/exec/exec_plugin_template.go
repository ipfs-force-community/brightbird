package main

import (
	"context"

	"github.com/hunjixin/brightbird/env/plugin"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
	"github.com/hunjixin/brightbird/env"
)

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

// 插件信息
var Info = types.PluginInfo{
	Name:        "plugin_name",  // 替换为你的插件名称
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "plugin_description",  // 替换为您的插件描述
}

// 在执行时需要使用的插件参数
type TestCaseParams struct {
    // 定义需要使用的插件参数
}

// 插件的执行逻辑
func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
    // 编写你的代码逻辑
	// 注意将单独的功能抽象为函数
    return nil
}
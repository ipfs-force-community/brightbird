// 测试插件编写原则：
// 1. 单一职责：每个插件尽量只测试一个单独的功能
// 2. 最小化原则：每个测试插件功能最小化
// 3. 参照模板：模板中{{}}标注的内容是需要手动填写的，其他框架部分无需修改

package main

import (
	"context"

	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"
)

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

// 插件信息
var Info = types.PluginInfo{
	Name:        "{{plugin_name}}", // 替换为你的插件名称
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "{{plugin_description}}", // 替换为您的插件描述
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

package damoclesmanager

import (
	"context"
	"embed"
	"fmt"

	"github.com/filecoin-project/go-address"
	"github.com/ipfs-force-community/brightbird/env"
	venusutils "github.com/ipfs-force-community/brightbird/env/venus_utils"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/utils"
	"github.com/ipfs-force-community/brightbird/version"
	"github.com/pelletier/go-toml"
	corev1 "k8s.io/api/core/v1"
)

type Config struct {
	env.BaseConfig
	VConfig
}

type VConfig struct {
	PieceStores   []string `jsonschema:"-"`
	PersistStores []string `jsonschema:"-"`

	NodeUrl      string `jsonschema:"-"`
	MessagerUrl  string `jsonschema:"-"`
	MarketUrl    string `jsonschema:"-"`
	GatewayUrl   string `jsonschema:"-"`
	MinerAddress string `jsonschema:"-"`

	SenderWalletAddress address.Address `json:"senderWalletAddress"  jsonschema:"senderWalletAddress" title:"SenderWalletAddress" require:"true" description:"sender wallet address"`
	UserToken           string          `json:"userToken" jsonschema:"userToken" title:"UserToken" require:"true" description:"user token"`
	SendFund            string          `jsonschema:"-"`
}

type DamoclesManagerReturn struct { //nolint
	VConfig
	env.CommonDeployParams
}
type RenderParams struct {
	Config

	NameSpace string
	Registry  string
	Args      []string

	UniqueId      string
	MountStorages []string // distinct with PieceStores and PersistStores
}

var PluginInfo = types.PluginInfo{
	Name:       "damocles-manager-exist",
	Version:    version.Version(),
	PluginType: types.Deploy,
	DeployPluginParams: types.DeployPluginParams{
		Repo:        "https://github.com/ipfs-force-community/damocles.git",
		ImageTarget: "damocles-manager",
		BuildScript: `sed -i "2 i\RUN sed -i 's/deb.debian.org/mirrors.ustc.edu.cn/g' /etc/apt/sources.list" Dockerfile.manager
sed -i "17 i\RUN go env -w GOPROXY=https://goproxy.cn,direct" Dockerfile.manager
sed -i '7 i\ENV HTTPS_PROXY="{{.Proxy}}"' Dockerfile.manager
sed -i '8 i\ENV RUSTUP_DIST_SERVER="https://rsproxy.cn"' Dockerfile.manager
sed -i '9 i\ENV RUSTUP_UPDATE_ROOT="https://rsproxy.cn/rustup"' Dockerfile.manager
sed -i "s/https:\/\/sh.rustup.rs/https:\/\/rsproxy.cn\/rustup-init.sh/g" Dockerfile.manager

sed -i '1 i\export GITHUB_TOKEN={{.GitToken}}' Makefile
sed -i '2 i\export HTTPS_PROXY={{.Proxy}}' Makefile

cat > ./config.toml << EOF
[source.crates-io]
replace-with = 'rsproxy'
[source.rsproxy]
registry = "https://rsproxy.cn/crates.io-index"
[source.rsproxy-sparse]
registry = "sparse+https://rsproxy.cn/index/"
[registries.rsproxy]
index = "https://rsproxy.cn/crates.io-index"
[net]
git-fetch-with-cli = true
EOF

cat > ./install-filcrypto << EOF
#!/usr/bin/env bash
# shellcheck disable=SC2155 enable=require-variable-braces

set -Exeo pipefail
auth_header=()
if [ -n "\${GITHUB_TOKEN}" ]; then
	auth_header=("-H" "Authorization: token \${GITHUB_TOKEN}")
fi

# set CWD to the root of filecoin-ffi
#
cd "\$(dirname "\${BASH_SOURCE[0]}")"

# tracks where the Rust sources are were we to build locally instead of
# downloading from GitHub Releases
#
rust_sources_dir="rust"

main() {
	download_release_tarball __tarball_path "\${rust_sources_dir}" "filecoin-ffi" ""
	local __tmp_dir=\$(mktemp -d)

	# silence shellcheck warning as the assignment happened in
	# 'download_release_tarball()'
	# shellcheck disable=SC2154
	# extract downloaded tarball to temporary directory
	#
	tar -C "\${__tmp_dir}" -xzf "\${__tarball_path}"

	# copy build assets into root of filecoin-ffi
	#

	find -L "\${__tmp_dir}" -type f -name filcrypto.h -exec cp -- "{}" . \;
	find -L "\${__tmp_dir}" -type f -name libfilcrypto.a -exec cp -- "{}" . \;
	find -L "\${__tmp_dir}" -type f -name filcrypto.pc -exec cp -- "{}" . \;

	check_installed_files

	(>&2 echo "[install-filcrypto/main] successfully installed prebuilt libfilcrypto")
}

download_release_tarball() {
	local __resultvar=\$1
	local __rust_sources_path=\$2
	local __repo_name=\$3
	local __release_flags=\$4
	local __release_sha1=\$(git rev-parse HEAD)
	local __release_tag="\${__release_sha1:0:16}"
	local __release_tag_url="https://api.github.com/repos/filecoin-project/\${__repo_name}/releases/tags/\${__release_tag}"

	# Download the non-optimized standard release.
	release_flag_name="standard"

	# TODO: This function shouldn't make assumptions about how these releases'
	# names are constructed. Marginally less-bad would be to require that this
	# function's caller provide the release name.
	#
	if [ "\$(uname -s)" = "Darwin" ]; then
		# For MacOS a universal library is used so naming convention is different
		local __release_name="\${__repo_name}-\$(uname)-\${release_flag_name}"
	else
		local __release_name="\${__repo_name}-\$(uname)-\$(uname -m)-\${release_flag_name}"
	fi

	(>&2 echo "[download_release_tarball] acquiring release @ \${__release_tag}")

	local __release_response=\$(curl "\${auth_header[@]}" \
		--retry 3 \
		--location "\${__release_tag_url}")

	local __release_url=\$(echo "\${__release_response}" | jq -r ".assets[] | select(.name | contains(\"\${__release_name}\")) | .url")

	local __tar_path="/tmp/\${__release_name}_\$(basename "\${__release_url}").tar.gz"

	if [[ -z "\${__release_url}" ]]; then
		(>&2 echo "[download_release_tarball] failed to download release (tag URL: \${__release_tag_url}, response: \${__release_response})")
		return 1
	fi

	local __asset_url=\$(curl "\${auth_header[@]}" \
		--head \
		--retry 3 \
		--header "Accept:application/octet-stream" \
		--location \
		--output /dev/null \
		-w "%{url_effective}" \
		"\${__release_url}")

	if ! curl --retry 3 --output "\${__tar_path}" "\${__asset_url}"; then
		(>&2 echo "[download_release_tarball] failed to download release asset (tag URL: \${__release_tag_url}, asset URL: \${__asset_url})")
		return 1
	fi

	# set \$__resultvar (which the caller provided as \$1), which is the poor
	# man's way of returning a value from a function in Bash
	#
	eval "\${__resultvar}='\${__tar_path}'"
}

check_installed_files() {
	pwd
	ls ./*filcrypto*

	if [[ ! -f "./filcrypto.h" ]]; then
		(>&2 echo "[check_installed_files] failed to install filcrypto.h")
		exit 1
	fi

	if [[ ! -f "./libfilcrypto.a" ]]; then
		(>&2 echo "[check_installed_files] failed to install libfilcrypto.a")
		exit 1
	fi

	if [[ ! -f "./filcrypto.pc" ]]; then
		(>&2 echo "[check_installed_files] failed to install filcrypto.pc")
		exit 1
	fi
}

main "\$@"; exit

EOF

git add config.toml install-filcrypto

sed -i '10 i\ENV RUSTFLAGS="-C target-cpu=x86-64"' Dockerfile.manager
sed -i '13 i\COPY ./config.toml /root/.cargo/config.toml' Dockerfile.manager
sed -i '14 i\ENV CARGO_HOME="/root/.cargo"' Dockerfile.manager
sed -i '32 i\RUN make -C damocles-manager/ build-dep/.update-modules' Dockerfile.manager
sed -i '33 i\RUN cp -f install-filcrypto damocles-manager/extern/filecoin-ffi/install-filcrypto' Dockerfile.manager

docker build -f Dockerfile.manager -t damocles-manager --build-arg HTTPS_PROXY={{.Proxy}} --build-arg FFI_BUILD_FROM_SOURCE=1 --build-arg FFI_USE_BLST_PORTABLE=1 --build-arg FFI_USE_OPENCL=1 .
docker tag damocles-manager {{.Registry}}/filvenus/damocles-manager:{{.Commit}}
docker push {{.Registry}}/filvenus/damocles-manager:{{.Commit}}`,
	},
	Description: "",
}

//go:embed damocles-manager-exist
var f embed.FS

func DeployFromConfig(ctx context.Context, k8sEnv *env.K8sEnvDeployer, cfg Config) (*DamoclesManagerReturn, error) {
	renderParams := RenderParams{
		NameSpace: k8sEnv.NameSpace(),
		Registry:  k8sEnv.Registry(),
		UniqueId:  env.UniqueId(k8sEnv.TestID(), k8sEnv.Retry(), cfg.InstanceName),
		Config:    cfg,
	}

	if utils.HasDupItemInArrary(renderParams.PieceStores) {
		return nil, fmt.Errorf("piece storage has same pvc %s", renderParams.PieceStores)
	}

	if utils.HasDupItemInArrary(renderParams.PersistStores) {
		return nil, fmt.Errorf("piece storage has same pvc %s", renderParams.PersistStores)
	}

	pvcFilter := make(map[string]bool)
	mountStorages := []string{}
	for _, name := range renderParams.PieceStores {
		_, ok := pvcFilter[name]
		if !ok {
			pvcFilter[name] = true
			mountStorages = append(mountStorages, name)
		}
	}
	for _, name := range renderParams.PersistStores {
		_, ok := pvcFilter[name]
		if !ok {
			pvcFilter[name] = true
			mountStorages = append(mountStorages, name)
		}
	}
	renderParams.MountStorages = mountStorages

	// create configMap
	configMapFs, err := f.Open("damocles-manager-exist/damocles-manager-exist-configmap.yaml")
	if err != nil {
		return nil, err
	}
	configMap, err := k8sEnv.RunConfigMap(ctx, configMapFs, renderParams)
	if err != nil {
		return nil, err
	}

	// create deployment
	deployCfg, err := f.Open("damocles-manager-exist/damocles-manager-exist-statefulset.yaml")
	if err != nil {
		return nil, err
	}
	statefulSet, err := k8sEnv.RunStatefulSets(ctx, func(ctx context.Context, k8sEnv *env.K8sEnvDeployer) ([]corev1.Pod, error) {
		return GetPods(ctx, k8sEnv, cfg.InstanceName)
	}, deployCfg, renderParams)
	if err != nil {
		return nil, err
	}

	// create service
	svcCfg, err := f.Open("damocles-manager-exist/damocles-manager-exist-headless.yaml")
	if err != nil {
		return nil, err
	}

	svc, err := k8sEnv.RunService(ctx, svcCfg, renderParams)
	if err != nil {
		return nil, err
	}

	svcEndpoint, err := k8sEnv.WaitForServiceReady(ctx, svc, venusutils.VenusHealthCheck)
	if err != nil {
		return nil, err
	}

	return &DamoclesManagerReturn{
		VConfig: cfg.VConfig,
		CommonDeployParams: env.CommonDeployParams{
			BaseConfig:      cfg.BaseConfig,
			DeployName:      PluginInfo.Name,
			StatefulSetName: statefulSet.GetName(),
			ConfigMapName:   configMap.GetName(),
			SVCName:         svc.GetName(),
			SvcEndpoint:     svcEndpoint,
		},
	}, nil
}

func GetConfig(ctx context.Context, k8sEnv *env.K8sEnvDeployer, configMapName string) (interface{}, error) {
	cfgData, err := k8sEnv.GetConfigMap(ctx, configMapName, "sector-manager.cfg")
	if err != nil {
		return nil, err
	}

	var cfg interface{}
	err = toml.Unmarshal(cfgData, &cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func Update(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params DamoclesManagerReturn, updateCfg interface{}) error {
	cfgData, err := toml.Marshal(updateCfg)
	if err != nil {
		return err
	}
	err = k8sEnv.SetConfigMap(ctx, params.ConfigMapName, "sector-manager.cfg", cfgData)
	if err != nil {
		return err
	}

	pods, err := GetPods(ctx, k8sEnv, params.InstanceName)
	if err != nil {
		return nil
	}
	for _, pod := range pods {
		_, err = k8sEnv.ExecRemoteCmd(ctx, pod.GetName(), "echo", "'"+string(cfgData)+"'", ">", "/root/.damocles-manager/sector-manager.cfg")
		if err != nil {
			return err
		}
	}

	err = k8sEnv.UpdateStatefulSetsByName(ctx, params.StatefulSetName)
	if err != nil {
		return err
	}
	return nil
}

func GetPods(ctx context.Context, k8sEnv *env.K8sEnvDeployer, instanceName string) ([]corev1.Pod, error) {
	return k8sEnv.GetPodsByLabel(ctx, fmt.Sprintf("damocles-manager-exist-%s-pod", env.UniqueId(k8sEnv.TestID(), k8sEnv.Retry(), instanceName)))
}

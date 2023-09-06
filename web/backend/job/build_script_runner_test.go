package job

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecScript_ExecScript222(t *testing.T) {
	runner := &ExecScript{
		PwdDir:   "/storage-nfs-4/li/buildspace/damocles",
		Proxy:    "http://192.168.200.34:7891",
		Registry: "http:/192.168.200.175",
		Env:      map[string]string{},
	}

	script := `sed -i "2 i\RUN sed -i 's/deb.debian.org/mirrors.ustc.edu.cn/g' /etc/apt/sources.list" Dockerfile.manager
sed -i "17 i\RUN go env -w GOPROXY=https://goproxy.cn,direct" Dockerfile.manager
sed -i '7 i\ENV RUSTUP_DIST_SERVER="https://rsproxy.cn"' Dockerfile.manager
sed -i '8 i\ENV RUSTUP_UPDATE_ROOT="https://rsproxy.cn/rustup"' Dockerfile.manager
sed -i "s/https:\/\/sh.rustup.rs/https:\/\/rsproxy.cn\/rustup-init.sh/g" Dockerfile.manager
cat > config << EOF
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

sed -i "12 i\COPY config /root/.cargo/config" Dockerfile.manager

sed -i "4 i\RUN sed -i 's/deb.debian.org/mirrors.ustc.edu.cn/g' /etc/apt/sources.list" damocles-worker/Dockerfile
sed -i "27 i\COPY config /root/.cargo/config" damocles-worker/Dockerfile
make docker-push TAG=a86638642c19edfe35cfc6b86a4c397d082e6c94 BUILD_DOCKER_PROXY=http://192.168.200.34:7891 PRIVATE_REGISTRY=192.168.200.175`

	fmt.Println(runner.ExecScript(context.Background(), BuildParams{
		Script: script,
		Commit: "test",
	}))

}
func TestExecScript_ExecScript(t *testing.T) {
	runner := &ExecScript{
		PwdDir:   ".",
		Proxy:    "http://127.0.0.1:7890",
		Registry: "http:/127.0.0.2:1234",
		Env:      map[string]string{},
	}

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		err := runner.ExecScript(ctx, BuildParams{
			Script: "ls",
			Commit: "",
		})
		assert.Nil(t, err)
	})

	t.Run("fail", func(t *testing.T) {
		err := runner.ExecScript(ctx, BuildParams{
			Script: "xx",
			Commit: "",
		})
		assert.Error(t, err)
	})

	t.Run("multipleline", func(t *testing.T) {
		script := `
echo {{.Proxy}}
echo {{.Commit}}
`
		err := runner.ExecScript(ctx, BuildParams{
			Script: script,
			Commit: "aaaa",
		})
		assert.Nil(t, err)
	})

	t.Run("multi runner", func(t *testing.T) {
		script := `
echo {{.Proxy}}
exit(1)
`
		for i := 0; i < 5; i++ {
			err := runner.ExecScript(ctx, BuildParams{
				Script: script,
				Commit: "aaaa",
			})
			assert.NotNil(t, err)
		}
	})
}

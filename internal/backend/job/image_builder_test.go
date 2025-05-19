package job

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetReoName(t *testing.T) {
	repoName, err := getRepoNameFromUrl("https://github.com/filecoin-project/venus-market.git")
	assert.Nil(t, err)
	assert.Equal(t, "venus-market", repoName)

	repoName, err = getRepoNameFromUrl("https://github.com/filecoin-project/venus-wallet.git")
	assert.Nil(t, err)
	assert.Equal(t, "venus-wallet", repoName)
}

func Test_TransferToSSH(t *testing.T) {
	repoName, err := toSSHFormat("https://github.com/filecoin-project/venus-market.git")
	assert.Nil(t, err)
	assert.Equal(t, "git@github.com:filecoin-project/venus-market.git", repoName)

	repoName, err = toSSHFormat("git@github.com:filecoin-project/venus-market.git")
	assert.Nil(t, err)
	assert.Equal(t, "git@github.com:filecoin-project/venus-market.git", repoName)
}

func Test_execCmd(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		err := execCmd(".", "/bin/sh", "-c", "ls")
		assert.Nil(t, err)
	})
	t.Run("fail", func(t *testing.T) {
		err := execCmd(".", "/bin/sh", "-c", "gg")
		assert.NotNil(t, err)
	})

	t.Run("multiple line", func(t *testing.T) {
		sh := `
ls
echo "hel"
`
		err := execCmd(".", "/bin/sh", "-c", sh)
		assert.Nil(t, err)
	})

	t.Run("multiple line fail", func(t *testing.T) {
		sh := `
ls
echo "hel"
exit(1)
`
		err := execCmd(".", "/bin/sh", "-c", sh)
		assert.NotNil(t, err)
	})
}

func Test_ToHttps(t *testing.T) {
	t.Run("correct", func(t *testing.T) {
		httpfmt, err := toHTTPFormat("git@github.com:ipfs-force-community/damocles.git")
		assert.Nil(t, err)
		assert.Equal(t, "https://github.com/ipfs-force-community/damocles.git", httpfmt)
	})

	t.Run("alread http", func(t *testing.T) {
		httpfmt, err := toHTTPFormat("https://github.com/ipfs-force-community/damocles.git")
		assert.Nil(t, err)
		assert.Equal(t, "https://github.com/ipfs-force-community/damocles.git", httpfmt)
	})
}

func Test_ToSSH(t *testing.T) {
	httpfmt, err := toSSHFormat("https://github.com/ipfs-force-community/damocles.git")
	assert.Nil(t, err)
	assert.Equal(t, "git@github.com:ipfs-force-community/damocles.git", httpfmt)
}

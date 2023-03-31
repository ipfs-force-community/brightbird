package job

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_uncompressFFI(t *testing.T) {
	uncompressFFI("/tmp/5ab95e3d9c5a9ceb791b485c301212ff7760af8c-filecoin-ffi-Linux-standard.tar.gz", "")
}

func Test_GetReoName(t *testing.T) {
	repoName, err := getRepoNameFromUrl("https://github.com/filecoin-project/venus-market.git")
	assert.Nil(t, err)
	assert.Equal(t, "venus-market", repoName)

	repoName, err = getRepoNameFromUrl("https://github.com/filecoin-project/venus-wallet.git")
	assert.Nil(t, err)
	assert.Equal(t, "venus-wallet", repoName)
}

func Test_TransferToSSH(t *testing.T) {
	repoName, err := toSShFormat("https://github.com/filecoin-project/venus-market.git")
	assert.Nil(t, err)
	assert.Equal(t, "git@github.com:filecoin-project/venus-market.git", repoName)

	repoName, err = toSShFormat("git@github.com:filecoin-project/venus-market.git")
	assert.Nil(t, err)
	assert.Equal(t, "git@github.com:filecoin-project/venus-market.git", repoName)
}

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

package env

import (
	"bytes"
	"io"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockFs struct {
	r io.Reader
}

func (m mockFs) Stat() (fs.FileInfo, error) {
	return nil, nil
}

func (m mockFs) Read(p []byte) (int, error) {
	return m.r.Read(p)
}

func (m mockFs) Close() error {
	return nil
}

func TestQuickRender(t *testing.T) {
	r := bytes.NewBufferString(`[{{if gt (len .BootstrapPeers) 0}}"{{join .BootstrapPeers "\",\""}}"{{end}}]`)
	data, err := QuickRender(mockFs{r: r}, map[string]interface{}{
		"BootstrapPeers": []string{"1", "2", "3"},
	})
	assert.Nil(t, err)
	assert.Equal(t, `["1","2","3"]`, string(data))
}

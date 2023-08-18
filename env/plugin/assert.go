package plugin

import (
	"fmt"

	"github.com/stretchr/testify/assert"
)

var Assert *assert.Assertions

func init() {
	Assert = assert.New(reportT{})
}

var _ assert.TestingT = (*reportT)(nil)

type reportT struct {
}

func (r reportT) Errorf(format string, args ...interface{}) {
	panic(fmt.Errorf(format, args...))
}

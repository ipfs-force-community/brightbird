package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCheckCount(t *testing.T) {
	t.SkipNow()

	endTime, err := time.Parse(timeFormat, "2023-09-17T07:43:21")
	assert.NoError(t, err)
	assert.NoError(t, checkActualComputeCount("sophon-miner.log", uint64(endTime.Unix()), 10, 30))
}

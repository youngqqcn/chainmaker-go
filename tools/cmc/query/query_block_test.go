package query

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetInputHeight(t *testing.T) {
	table := []struct {
		args     []string
		height   uint64
		hasError bool
	}{
		{
			args:     []string{"123"},
			height:   123,
			hasError: false,
		},
		{
			args:     []string{"-1"},
			height:   math.MaxUint64,
			hasError: false,
		},
		{
			args:     []string{"123ff"},
			height:   0,
			hasError: true,
		},
		{
			args:     []string{},
			height:   0,
			hasError: true,
		},
	}
	for _, row := range table {
		h, e := getInputHeight(row.args)
		assert.Equal(t, row.height, h, row.args)
		if row.hasError {
			assert.NotNil(t, e, row.args)
		} else {
			assert.NoError(t, e, row.args)
		}
	}
}

package cmd_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func assertEqualContentFile(t *testing.T, filepath1, filepath2 string) {
	contents1, err := os.ReadFile(filepath1)
	assert.NoError(t, err)

	contents2, err := os.ReadFile(filepath2)
	assert.NoError(t, err)

	assert.Equal(t, string(contents1), string(contents2))
}

func ptStr(s string) *string {
	return &s
}

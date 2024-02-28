package utils_test

import (
	"testing"

	"github.com/happyagosmith/jig/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestExtractValue(t *testing.T) {
	tests := []struct {
		name     string
		yamlData []byte
		path     string
		expected string
	}{
		{
			name: "Test 1",
			yamlData: []byte(`
a:
  b:
    c: hello
`),
			path:     "$.a.b.c",
			expected: "hello",
		},
		{
			name: "Test 2",
			yamlData: []byte(`
a:
  - b: hello
`),
			path:     "$.a[0].b",
			expected: "hello",
		},
		{
			name: "Test 2",
			yamlData: []byte(`
a:
  - b: hello
    c: label
`),
			path:     "$.a[?(@.c == 'label')].b",
			expected: "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := utils.GetYamlValue(tt.yamlData, tt.path)
			assert.NoError(t, err, "not expected error")
			if result != tt.expected {
				t.Errorf("ExtractValue(%s) = %s; want %s", tt.path, result, tt.expected)
			}
		})
	}
}

package yaml_test

import (
	"testing"

	"github.com/happyagosmith/jig/internal/filehandler/yaml"
	"github.com/stretchr/testify/assert"
)

func TestYaml(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		want := "key1: v1\n" +
			"key2:\n" +
			"  key2-1: v2-1\n" +
			"  key2-2: v2-2\n" +
			"key3:\n" +
			"  - a\n" +
			"  - b\n" +
			"key4:\n" +
			"  - t1: 1\n" +
			"    t2: 2"
		y, _ := yaml.NewYaml([]byte(want))
		got, _ := y.String()
		assert.Equal(t, want, got, "output yaml")
	})
}

func TestMergeYaml(t *testing.T) {
	type testCase struct {
		name string
		y1   string
		y2   string
		want string
	}

	tests := []testCase{
		{
			name: "same key",
			y1:   "key1: v1",
			y2:   "key1: v1-overwritten",
			want: "key1: v1-overwritten",
		},
		{
			name: "same key in map",
			y1: "key4:\n" +
				"  key4-1: v4-1\n" +
				"  key4-2: v4-2",
			y2: "key4:\n" +
				"  key4-1: v4-overwritten\n" +
				"  key4-2: v4-2",
			want: "key4:\n" +
				"  key4-1: v4-overwritten\n" +
				"  key4-2: v4-2",
		},
		{
			name: "different keys in map",
			y1: "key4:\n" +
				"  key4-2: v4-1",
			y2: "key4:\n" +
				"  key4-1: v4-2",
			want: "key4:\n" +
				"  key4-2: v4-1\n" +
				"  key4-1: v4-2",
		},
		{
			name: "list",
			y1: "key3:\n" +
				"  - a\n" +
				"  - b",
			y2: "key3:\n" +
				"  - c",
			want: "key3:\n" +
				"  - c",
		},
		{
			name: "list of map",
			y1: "key4:\n" +
				"  - t1: 1\n" +
				"    t2: 2",
			y2: "key4:\n" +
				"  - t1: 2\n" +
				"    t3: 2",
			want: "key4:\n" +
				"  - t1: 2\n" +
				"    t3: 2",
		},
		{
			name: "generate sorted yaml",
			y1: "key2: v2\n" +
				"key3: v3\n" +
				"key4: v4",
			y2: "key1: v1\n" +
				"key2: v2-overwritten\n" +
				"key4: v4-overwritten",
			want: "key2: v2-overwritten\n" +
				"key3: v3\n" +
				"key4: v4-overwritten\n" +
				"key1: v1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ny1, _ := yaml.NewYaml([]byte(tt.y1))
			ny2, _ := yaml.NewYaml([]byte(tt.y2))
			ny1.Merge(ny2)

			got, _ := ny1.String()
			assert.Equal(t, tt.want, got, "merged yaml")
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		key         string
		wantErr     bool
		wantContent string
	}{
		{
			name: "delete existing key",
			content: `key1: value1
key2: value2
`,
			key:         "key1",
			wantErr:     false,
			wantContent: "key2: value2\n",
		},
		{
			name: "delete non-existing key",
			content: `key1: value1
key2: value2
`,
			key:     "key3",
			wantErr: false,
			wantContent: `key1: value1
key2: value2
`,
		},
		{
			name:        "delete empty content",
			content:     "",
			key:         "key1",
			wantErr:     false,
			wantContent: "",
		},
		{
			name:        "delete unique key",
			content:     `key1: value1`,
			key:         "key1",
			wantErr:     false,
			wantContent: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			y, err := yaml.NewYaml([]byte(tt.content))
			if err != nil {
				t.Fatalf("NewYaml() error = %v", err)
			}

			err = y.Delete(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			bytes, err := y.Bytes()
			if err != nil {
				t.Fatalf("Bytes() error = %v", err)
			}

			if string(bytes) != tt.wantContent {
				t.Errorf("Delete() = %v, want %v", string(bytes), tt.wantContent)
			}
		})
	}
}

func TestGetValue(t *testing.T) {
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
			name: "Test 3",
			yamlData: []byte(`
a:
  - b: hello
    c: label
`),
			path:     "$.a[?(@.c == 'label')].b",
			expected: "hello",
		},
		{
			name: "Test 4",
			yamlData: []byte(`
a: hello
`),
			path:     "a",
			expected: "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			y, err := yaml.NewYaml([]byte(tt.yamlData))
			if err != nil {
				t.Fatalf("NewYaml() error = %v", err)
			}

			result, err := y.GetValue(tt.path)
			assert.NoError(t, err, "not expected error")
			if result != tt.expected {
				t.Errorf("ExtractValue(%s) = %s; want %s", tt.path, result, tt.expected)
			}
		})
	}
}

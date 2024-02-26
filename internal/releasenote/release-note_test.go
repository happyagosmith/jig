package releaseNote_test

import (
	"bytes"
	"os"
	"testing"

	releaseNote "github.com/happyagosmith/jig/internal/releasenote"
	"gopkg.in/yaml.v2"
)

func TestGenerate(t *testing.T) {
	tests := []struct {
		name       string
		tplContent string
		issues     map[any]any
		expected   string
	}{
		{
			name:       "Test with service keys",
			tplContent: "{{range $key, $value := . }}{{ $key }},{{end}}",
			issues: map[any]any{
				"service1": []map[any]any{{"issueKey": "1"}, {"issueKey": "2"}},
				"service2": []map[any]any{{"issueKey": "1"}, {"issueKey": "3"}},
			},
			expected: "service1,service2,",
		},
		{
			name:       "Test with issuesFlatList",
			tplContent: `{{range (issuesFlatList .)}} {{ .issueKey}}: {{range .impactedService}}{{.}},{{end}}{{end}}`,
			issues: map[any]any{
				"service1": []map[any]any{{"issueKey": "1"}, {"issueKey": "2"}},
				"service2": []map[any]any{{"issueKey": "1"}, {"issueKey": "3"}},
			},
			expected: " 1: service1,service2, 2: service1, 3: service2,",
		},
		{
			name:       "Test with issuesFlatList and join function",
			tplContent: `{{range (issuesFlatList .)}} {{ .issueKey}}: {{ .impactedService | join ","}}{{end}}`,
			issues: map[any]any{
				"service1": []map[any]any{{"issueKey": "1"}, {"issueKey": "2"}},
				"service2": []map[any]any{{"issueKey": "1"}, {"issueKey": "3"}},
			},
			expected: " 1: service1,service2 2: service1 3: service2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, err := yaml.Marshal(tt.issues)
			if err != nil {
				t.Fatal(err)
			}

			tplFile, err := os.CreateTemp("", "template")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tplFile.Name())

			tplFile.WriteString(tt.tplContent)
			tplFile.Close()

			mockWriter := &bytes.Buffer{}
			err = releaseNote.Generate(tplFile.Name(), values, mockWriter)
			if err != nil {
				t.Fatal(err)
			}

			if mockWriter.String() != tt.expected {
				t.Errorf("Expected %s, but got %s", tt.expected, mockWriter.String())
			}
		})
	}
}

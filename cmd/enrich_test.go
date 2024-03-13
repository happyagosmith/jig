package cmd_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/happyagosmith/jig/cmd"
	shell "github.com/mattn/go-shellwords"
	"github.com/stretchr/testify/assert"
)

type testCase struct {
	name           string
	mockgitcommits string
	mockgitissues  string
	mockjira       string
	model          string
	cmd            string
	wantModel      string
	wantRN         string
}

func TestEnrich(t *testing.T) {
	tests := []testCase{
		{
			name:           "test enrich",
			mockgitcommits: "testdata/gitlab-compare.json",
			mockgitissues:  "testdata/gitlab-issues.json",
			mockjira:       "testdata/jira-issues.json",
			model:          "testdata/model.yaml",
			cmd:            "enrich %s",
			wantModel:      "testdata/want-model.yaml",
			wantRN:         "testdata/want-rn.md",
		},
	}

	runTests(t, tests)
}

func runTests(t *testing.T, tests []testCase) {
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotJiraRequest *http.Request
			jirasrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotJiraRequest = r
				w.WriteHeader(200)
				b, err := os.ReadFile(tt.mockjira)
				if err != nil {
					t.Fatal(err)
				}

				w.Write(b)
			}))
			defer jirasrv.Close()

			var gotGitRequest []*http.Request
			gitsrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotGitRequest = append(gotGitRequest, r)
				if r.URL.Path == "/api/v4/projects/123/repository/compare" {
					w.WriteHeader(200)
					b, err := os.ReadFile(tt.mockgitcommits)
					if err != nil {
						t.Fatal(err)
					}
					w.Write(b)
				} else if r.URL.Path == "/api/v4/projects/123/issues" {
					w.WriteHeader(200)
					b, err := os.ReadFile(tt.mockgitissues)
					if err != nil {
						t.Fatal(err)
					}
					w.Write(b)
				} else {
					http.Error(w, "Not found", http.StatusNotFound)
				}
			}))
			defer gitsrv.Close()

			modelFile, err := os.CreateTemp("", "model.yaml")
			assert.NoError(t, err)
			defer os.Remove(modelFile.Name())

			f, err := os.ReadFile(tt.model)
			assert.NoError(t, err)

			_, err = modelFile.Write(f)
			assert.NoError(t, err)

			cmdline := fmt.Sprintf(tt.cmd+" --jiraURL %s --gitURL %s", modelFile.Name(), jirasrv.URL, gitsrv.URL)
			args, _ := shell.Parse(cmdline)
			rootCmd := cmd.NewRootCmd("0.0.1")
			rootCmd.SetArgs(args)
			err = rootCmd.Execute()

			assert.NoError(t, err)
			assert.NotNil(t, gotJiraRequest)
			assert.Equal(t, 2, len(gotGitRequest))

			asserEqualFile(t, tt.wantModel, modelFile.Name())
		})
	}
}

func asserEqualFile(t *testing.T, file1, file2 string) {
	contents1, err := os.ReadFile(file1)
	assert.NoError(t, err)

	contents2, err := os.ReadFile(file2)
	assert.NoError(t, err)

	assert.Equal(t, string(contents1), string(contents2))
}

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

func TestEnrich(t *testing.T) {
	tests := []struct {
		name           string
		mockgitcommits string
		mockgitissues  string
		mockgitmr      string
		mockjira       string
		model          string
		cmd            string
		wantModel      string
	}{
		{
			name:           "test enrich",
			mockgitcommits: "testdata/gitlab-compare.json",
			mockgitissues:  "testdata/gitlab-issues.json",
			mockgitmr:      "testdata/gitlab-mergerequest.json",
			mockjira:       "testdata/jira-issues.json",
			model:          "testdata/model.yaml",
			cmd:            "enrich %s",
			wantModel:      "testdata/model-enriched.yaml",
		},
	}

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
				} else if r.URL.Path == "/api/v4/projects/123/merge_requests" {
					w.WriteHeader(200)
					b, err := os.ReadFile(tt.mockgitmr)
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

			cmdline := fmt.Sprintf(tt.cmd+"  --gitMRBranch main --config testdata/config.yaml --jiraURL %s --gitURL %s", modelFile.Name(), jirasrv.URL, gitsrv.URL)
			fmt.Println("running: " + cmdline)

			args, _ := shell.Parse(cmdline)
			rootCmd := cmd.NewRootCmd("0.0.1")
			rootCmd.SetArgs(args)
			err = rootCmd.Execute()

			assert.NoError(t, err)
			assert.NotNil(t, gotJiraRequest)
			assert.Equal(t, 3, len(gotGitRequest))

			assertEqualContentFile(t, tt.wantModel, modelFile.Name())
		})
	}
}

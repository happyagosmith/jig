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

func TestGenerate(t *testing.T) {
	tests := []struct {
		name           string
		mockgitcommits *string
		mockgitissues  *string
		mockgitmr      *string
		mockjira       *string
		model          string
		rntpl          string
		cmd            string
		wantrn         string
	}{
		{
			name:           "test generate",
			mockgitcommits: nil,
			mockgitissues:  nil,
			mockgitmr:      nil,
			mockjira:       nil,
			model:          "testdata/model-enriched.yaml",
			rntpl:          "testdata/rn.tpl",
			cmd:            "generate %s -m %s",
			wantrn:         "testdata/rn.md",
		},
		{
			name:           "test generate with enrich",
			mockgitcommits: ptStr("testdata/gitlab-compare.json"),
			mockgitissues:  ptStr("testdata/gitlab-issues.json"),
			mockgitmr:      ptStr("testdata/gitlab-mergerequest.json"),
			mockjira:       ptStr("testdata/jira-issues.json"),
			model:          "testdata/model-enriched.yaml",
			rntpl:          "testdata/rn.tpl",
			cmd:            "generate --withEnrich %s -m %s",
			wantrn:         "testdata/rn.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotJiraRequest *http.Request
			jirasrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotJiraRequest = r
				w.WriteHeader(200)
				b, err := os.ReadFile(*tt.mockjira)
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
					b, err := os.ReadFile(*tt.mockgitcommits)
					if err != nil {
						t.Fatal(err)
					}
					w.Write(b)
				} else if r.URL.Path == "/api/v4/projects/123/issues" {
					w.WriteHeader(200)
					b, err := os.ReadFile(*tt.mockgitissues)
					if err != nil {
						t.Fatal(err)
					}
					w.Write(b)
				} else if r.URL.Path == "/api/v4/projects/123/merge_requests" {
					w.WriteHeader(200)
					b, err := os.ReadFile(*tt.mockgitmr)
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

			outputFile, err := os.CreateTemp("", "rn.md")
			assert.NoError(t, err)
			defer os.Remove(outputFile.Name())

			cmdline := fmt.Sprintf(tt.cmd+" --gitMRBranch main --config testdata/config.yaml  -o %s --jiraURL %s --gitURL %s", tt.rntpl, modelFile.Name(), outputFile.Name(), jirasrv.URL, gitsrv.URL)
			fmt.Println("running: " + cmdline)
			args, _ := shell.Parse(cmdline)
			rootCmd := cmd.NewRootCmd("0.0.1")
			rootCmd.SetArgs(args)
			err = rootCmd.Execute()

			assert.NoError(t, err)
			if tt.mockgitcommits != nil && tt.mockgitissues != nil {
				assert.Equal(t, 3, len(gotGitRequest))
			}

			if tt.mockjira != nil {
				assert.NotNil(t, gotJiraRequest)
			}

			assertEqualContentFile(t, tt.wantrn, outputFile.Name())
		})
	}
}

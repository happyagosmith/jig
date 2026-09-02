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

func TestSet(t *testing.T) {
	tests := []struct {
		name      string
		model     string
		cmd       string
		wantModel string
	}{
		{
			name:      "test setVersions",
			model:     "testdata/model.yaml",
			cmd:       "setVersions %s",
			wantModel: "testdata/model-updated.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			gitsrv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				if req.URL.Path == "/api/v4/projects/123/releases/0.0.3" {
					rw.Write([]byte(`{"_links": { "self": "https://jig-test-url/-/releases/0.0.3"}}`))
				} else if req.URL.Path == "/api/v4/projects/123" {
					rw.Write([]byte(`{"web_url": "https://jig-test-url/repo"}`))
				} else {
					http.Error(rw, "Not found", http.StatusNotFound)
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

			cmdline := fmt.Sprintf(tt.cmd+"  --config testdata/config.yaml --gitURL %s", modelFile.Name(), gitsrv.URL)
			fmt.Println("running: " + cmdline)
			args, _ := shell.Parse(cmdline)
			rootCmd := cmd.NewRootCmd("0.0.1")
			rootCmd.SetArgs(args)
			err = rootCmd.Execute()
			assert.NoError(t, err)

			assertEqualContentFile(t, tt.wantModel, modelFile.Name())
		})
	}
}

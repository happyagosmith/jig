package repoclients_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/happyagosmith/jig/internal/repoclients"
	"github.com/stretchr/testify/assert"
)

func TestGetRepoURL(t *testing.T) {
	gitRepoID := "123"

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/api/v4/projects/"+gitRepoID {
			rw.Write([]byte(`{"web_url": "https://gitlab.example.com/my/repo"}`))
		} else {
			http.Error(rw, "Not found", http.StatusNotFound)
		}
	}))

	g, err := repoclients.NewGitLab(server.URL, "token")
	assert.NoError(t, err)
	releaseURL, err := g.GetRepoURL(gitRepoID)

	assert.NoError(t, err)
	assert.Equal(t, "https://gitlab.example.com/my/repo", releaseURL)
}

func TestGetReleaseURL(t *testing.T) {
	gitRepoID := "123"
	version := "v1.0.0"

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == fmt.Sprintf("/api/v4/projects/%s/releases/%s", gitRepoID, version) {
			rw.Write([]byte(`{"_links": { "self": "https://gitlab.example.com/my/repo/releases/v1.0.0"}}`))
		} else {
			http.Error(rw, "Not found", http.StatusNotFound)
		}
	}))

	g, err := repoclients.NewGitLab(server.URL, "token")
	assert.NoError(t, err)
	releaseURL, err := g.GetReleaseURL(gitRepoID, version)

	assert.NoError(t, err)
	assert.Equal(t, "https://gitlab.example.com/my/repo/releases/v1.0.0", releaseURL)
}

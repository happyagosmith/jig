package git_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/happyagosmith/jig/internal/git"
	"github.com/happyagosmith/jig/internal/parsers"
	"github.com/stretchr/testify/assert"
)

func TestGitCommit(t *testing.T) {
	t.Run("parse jira commits", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			b, _ := os.ReadFile("testdata/git-compare.json")
			w.Write(b)
		}))
		defer srv.Close()

		gc, err := git.NewClient(srv.URL, "token", `(?P<jira_1>[^\]]*)`, git.WithCustomPattern(`\[(?P<scope>[^\]]*)\](?P<subject>.*)`))
		assert.NoError(t, err, "NewGit error must be nil")

		i, err := gc.ExtractCommits("", "from", "to")
		assert.NoError(t, err, "ExtractKeys error must be nil")

		assert.Equal(t, "AAA-1234", i[0].IssueKey)
		assert.Equal(t, parsers.JIRA, i[0].IssueTracker)
		assert.Equal(t, false, i[0].IsBreaking)
		assert.Equal(t, parsers.UNKNOWN, i[0].Category)
		assert.Equal(t, "With reference", i[0].Summary)

		assert.Equal(t, "AAA-5678", i[1].IssueKey)
		assert.Equal(t, parsers.JIRA, i[1].IssueTracker)
		assert.Equal(t, false, i[1].IsBreaking)
		assert.Equal(t, parsers.UNKNOWN, i[1].Category)
		assert.Equal(t, "Different reference", i[1].Summary)
	})

	t.Run("parse feature from conventional commit", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			b, _ := os.ReadFile("testdata/git-compare-conventional-commit.json")
			w.Write(b)
		}))
		defer srv.Close()

		gc, err := git.NewClient(srv.URL, "token", `j_(?P<jira_1>.*)`)
		assert.NoError(t, err, "NewGit error must be nil")

		i, err := gc.ExtractCommits("", "from", "to")
		assert.NoError(t, err, "ParseCommits error must be nil")

		assert.Equal(t, "CC-123", i[0].IssueKey)
		assert.Equal(t, parsers.JIRA, i[0].IssueTracker)
		assert.Equal(t, false, i[0].IsBreaking)
		assert.Equal(t, parsers.FEATURE, i[0].Category)
		assert.Equal(t, "this is a feature tracked in jira", i[0].Summary)
	})

	t.Run("parse bug fixed from conventional commit", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			b, _ := os.ReadFile("testdata/git-compare-conventional-commit.json")
			w.Write(b)
		}))
		defer srv.Close()

		gc, err := git.NewClient(srv.URL, "token", `j_(?P<jira_1>.*)`)
		assert.NoError(t, err, "NewGit error must be nil")

		i, err := gc.ExtractCommits("", "from", "to")
		assert.NoError(t, err, "ParseCommits error must be nil")

		assert.Equal(t, "CC-456", i[1].IssueKey)
		assert.Equal(t, parsers.JIRA, i[1].IssueTracker)
		assert.Equal(t, false, i[1].IsBreaking)
		assert.Equal(t, parsers.BUG_FIX, i[1].Category)
		assert.Equal(t, "this is a bug fixed tracked in jira", i[1].Summary)
	})

	t.Run("parse breaking change from conventional commit", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			b, _ := os.ReadFile("testdata/git-compare-conventional-commit.json")
			w.Write(b)
		}))
		defer srv.Close()

		gc, err := git.NewClient(srv.URL, "token", `j_(?P<jira_1>.*)`)
		assert.NoError(t, err, "NewGit error must be nil")

		i, err := gc.ExtractCommits("", "from", "to")
		assert.NoError(t, err, "ParseCommits error must be nil")

		assert.Equal(t, true, i[2].IsBreaking)
	})

	t.Run("parse unknown issue tracker from conventional commit", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			b, _ := os.ReadFile("testdata/git-compare-conventional-commit.json")
			w.Write(b)
		}))
		defer srv.Close()

		gc, err := git.NewClient(srv.URL, "token", `j_(?P<jira_1>.*)`)
		assert.NoError(t, err, "NewGit error must be nil")

		i, err := gc.ExtractCommits("", "from", "to")
		assert.NoError(t, err, "ParseCommits error must be nil")

		assert.Equal(t, "CC-222", i[3].IssueKey)
		assert.Equal(t, parsers.NONE, i[3].IssueTracker)
		assert.Equal(t, false, i[3].IsBreaking)
		assert.Equal(t, parsers.BUG_FIX, i[3].Category)
		assert.Equal(t, "this has an unknown issue tracker", i[3].Summary)
	})

	t.Run("add feature without issuekey from conventional commit", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			b, _ := os.ReadFile("testdata/git-compare-conventional-commit.json")
			w.Write(b)
		}))
		defer srv.Close()

		gc, err := git.NewClient(srv.URL, "token", `j_(?P<jira_1>.*)`, git.WithKeepCCWithoutScope(true))
		assert.NoError(t, err, "NewGit error must be nil")

		i, err := gc.ExtractCommits("", "from", "to")
		assert.NoError(t, err, "ParseCommits error must be nil")

		assert.Equal(t, "", i[4].IssueKey)
		assert.Equal(t, parsers.NONE, i[4].IssueTracker)
		assert.Equal(t, false, i[4].IsBreaking)
		assert.Equal(t, parsers.BUG_FIX, i[4].Category)
		assert.Equal(t, "this has an unknown issue tracker and no issueKey", i[4].Summary)
	})

	t.Run("not add feature without issuekey from conventional commit", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			b, _ := os.ReadFile("testdata/git-compare-conventional-commit.json")
			w.Write(b)
		}))
		defer srv.Close()

		gc, err := git.NewClient(srv.URL, "token", `j_(?P<jira_1>.*)`)
		assert.NoError(t, err, "NewGit error must be nil")

		i, err := gc.ExtractCommits("", "from", "to")
		assert.NoError(t, err, "ParseCommits error must be nil")

		assert.Equal(t, 4, len(i))
	})

	t.Run("test api error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(404)
		}))
		defer srv.Close()

		git, err := git.NewClient(srv.URL, "token", `\[([^\]]*)\]`)
		assert.NoError(t, err, "NewGit error must be nil")

		_, err = git.ExtractCommits("", "from", "to")
		assert.NotNil(t, err, "ParseCommits should return error")
	})
}

func TestGetRepoURL(t *testing.T) {
	gitRepoID := "123"

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/api/v4/projects/"+gitRepoID {
			rw.Write([]byte(`{"web_url": "https://gitlab.example.com/my/repo"}`))
		} else {
			http.Error(rw, "Not found", http.StatusNotFound)
		}
	}))

	g, err := git.NewClient(server.URL, "token", "")
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

	g, err := git.NewClient(server.URL, "token", "")
	assert.NoError(t, err)
	releaseURL, err := g.GetReleaseURL(gitRepoID, version)

	assert.NoError(t, err)
	assert.Equal(t, "https://gitlab.example.com/my/repo/releases/v1.0.0", releaseURL)
}

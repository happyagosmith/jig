package git_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/happyagosmith/jig/internal/git"
	"github.com/happyagosmith/jig/internal/parsers"
	"github.com/stretchr/testify/assert"
)

func TestGit(t *testing.T) {
	t.Run("parse jira commits", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			b, _ := os.ReadFile("data/git-compare.json")
			w.Write(b)
		}))
		defer srv.Close()

		gc, err := git.NewClient(srv.URL, "token", `(?P<jira_1>[^\]]*)`, "")
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
			b, _ := os.ReadFile("data/git-compare-conventional-commit.json")
			w.Write(b)
		}))
		defer srv.Close()

		gc, err := git.NewClient(srv.URL, "token", `j_(?P<jira_1>.*)`, "")
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
			b, _ := os.ReadFile("data/git-compare-conventional-commit.json")
			w.Write(b)
		}))
		defer srv.Close()

		gc, err := git.NewClient(srv.URL, "token", `j_(?P<jira_1>.*)`, "")
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
			b, _ := os.ReadFile("data/git-compare-conventional-commit.json")
			w.Write(b)
		}))
		defer srv.Close()

		gc, err := git.NewClient(srv.URL, "token", `j_(?P<jira_1>.*)`, "")
		assert.NoError(t, err, "NewGit error must be nil")

		i, err := gc.ExtractCommits("", "from", "to")
		assert.NoError(t, err, "ParseCommits error must be nil")

		assert.Equal(t, true, i[2].IsBreaking)
	})

	t.Run("parse unknown issue tracker from conventional commit", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			b, _ := os.ReadFile("data/git-compare-conventional-commit.json")
			w.Write(b)
		}))
		defer srv.Close()

		gc, err := git.NewClient(srv.URL, "token", `j_(?P<jira_1>.*)`, "")
		assert.NoError(t, err, "NewGit error must be nil")

		i, err := gc.ExtractCommits("", "from", "to")
		assert.NoError(t, err, "ParseCommits error must be nil")

		assert.Equal(t, "CC-222", i[3].IssueKey)
		assert.Equal(t, parsers.NONE, i[3].IssueTracker)
		assert.Equal(t, false, i[3].IsBreaking)
		assert.Equal(t, parsers.BUG_FIX, i[3].Category)
		assert.Equal(t, "this has an unknown issue tracker", i[3].Summary)
	})

	t.Run("test api error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(404)
		}))
		defer srv.Close()

		git, err := git.NewClient(srv.URL, "token", `\[([^\]]*)\]`, "")
		assert.NoError(t, err, "NewGit error must be nil")

		_, err = git.ExtractCommits("", "from", "to")
		assert.NotNil(t, err, "ParseCommits should return error")
	})
}

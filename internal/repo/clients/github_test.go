package clients_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/happyagosmith/jig/internal/entities"
	"github.com/happyagosmith/jig/internal/repo/clients"
	"github.com/stretchr/testify/assert"
)

func newGitHubTestServer(mux *http.ServeMux) (*httptest.Server, func()) {
	srv := httptest.NewServer(mux)
	return srv, srv.Close
}

func TestGitHubGetRepoURL(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v3/repos/owner/repo", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"html_url": "https://github.com/owner/repo"})
	})
	srv, stop := newGitHubTestServer(mux)
	defer stop()

	g, err := clients.NewGitHubWithBaseURL("token", srv.URL+"/api/v3/")
	assert.NoError(t, err)

	url, err := g.GetRepoURL("owner/repo")
	assert.NoError(t, err)
	assert.Equal(t, "https://github.com/owner/repo", url)
}

func TestGitHubGetReleaseURL(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v3/repos/owner/repo/releases/tags/v1.0.0", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"html_url": "https://github.com/owner/repo/releases/tag/v1.0.0"})
	})
	srv, stop := newGitHubTestServer(mux)
	defer stop()

	g, err := clients.NewGitHubWithBaseURL("token", srv.URL+"/api/v3/")
	assert.NoError(t, err)

	url, err := g.GetReleaseURL("owner/repo", "v1.0.0")
	assert.NoError(t, err)
	assert.Equal(t, "https://github.com/owner/repo/releases/tag/v1.0.0", url)
}

func TestGitHubGetCommits(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v3/repos/owner/repo/compare/v0.0.0...v0.0.1", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"commits": []map[string]any{
				{
					"sha":      "abc1234567890",
					"html_url": "https://github.com/owner/repo/commit/abc1234567890",
					"commit": map[string]any{
						"message": "Test commit\nBody of the commit",
						"committer": map[string]any{
							"date": "2021-01-01T00:00:00Z",
						},
					},
				},
			},
		})
	})
	srv, stop := newGitHubTestServer(mux)
	defer stop()

	g, err := clients.NewGitHubWithBaseURL("token", srv.URL+"/api/v3/")
	assert.NoError(t, err)

	commits, err := g.GetCommits("owner/repo", "v0.0.0", "v0.0.1")
	assert.NoError(t, err)
	assert.Len(t, commits, 1)

	wantCreatedAt := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, "abc1234567890", commits[0].ID)
	assert.Equal(t, "abc12345", commits[0].ShortID)
	assert.Equal(t, "Test commit", commits[0].Title)
	assert.Equal(t, "Test commit\nBody of the commit", commits[0].Message)
	assert.Equal(t, &wantCreatedAt, commits[0].CreatedAt)
	assert.Equal(t, "https://github.com/owner/repo/commit/abc1234567890", commits[0].WebURL)
	assert.Equal(t, "commit", commits[0].Origin)
}

func TestGitHubGetMergeRequests(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v3/repos/owner/repo/pulls", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "closed", r.URL.Query().Get("state"))
		assert.Equal(t, "main", r.URL.Query().Get("base"))
		mergedAt := "2021-01-01T00:00:00Z"
		json.NewEncoder(w).Encode([]map[string]any{
			{
				"id":               10,
				"number":           1,
				"title":            "fix: something",
				"body":             "PR description",
				"merge_commit_sha": "commitSHA1",
				"html_url":         "https://github.com/owner/repo/pull/1",
				"merged_at":        mergedAt,
			},
			{
				"id":               20,
				"number":           2,
				"title":            "feat: other",
				"body":             "another PR",
				"merge_commit_sha": "commitSHA2",
				"html_url":         "https://github.com/owner/repo/pull/2",
				"merged_at":        mergedAt,
			},
		})
	})
	srv, stop := newGitHubTestServer(mux)
	defer stop()

	g, err := clients.NewGitHubWithBaseURL("token", srv.URL+"/api/v3/")
	assert.NoError(t, err)

	wantCreatedAt := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	commits := []entities.RepoRecord{
		{ID: "commitSHA1", CreatedAt: &wantCreatedAt},
	}

	mrs, err := g.GetMergeRequests("owner/repo", "main", commits)
	assert.NoError(t, err)
	assert.Len(t, mrs, 1)
	assert.Equal(t, "10", mrs[0].ID)
	assert.Equal(t, "1", mrs[0].ShortID)
	assert.Equal(t, "fix: something", mrs[0].Title)
	assert.Equal(t, "PR description", mrs[0].Message)
	assert.Equal(t, &wantCreatedAt, mrs[0].CreatedAt)
	assert.Equal(t, "https://github.com/owner/repo/pull/1", mrs[0].WebURL)
	assert.Equal(t, "merge_request", mrs[0].Origin)
}

func TestGitHubGetIssues(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v3/repos/owner/repo/issues/42", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"number":   42,
			"title":    "Some bug",
			"state":    "closed",
			"html_url": "https://github.com/owner/repo/issues/42",
			"labels":   []map[string]any{{"name": "bug"}},
		})
	})
	srv, stop := newGitHubTestServer(mux)
	defer stop()

	g, err := clients.NewGitHubWithBaseURL("token", srv.URL+"/api/v3/")
	assert.NoError(t, err)

	repo := &entities.EnrichedRepo{Repo: entities.Repo{ID: "owner/repo"}}
	issues, err := g.GetIssues(context.Background(), repo, []string{"42"})
	assert.NoError(t, err)
	assert.Len(t, issues, 1)
	assert.Equal(t, "42", issues[0].IssueKey)
	assert.Equal(t, "Some bug", issues[0].IssueSummary)
	assert.Equal(t, "closed", issues[0].IssueStatus)
	assert.Equal(t, entities.FIXED_BUG, issues[0].Category)
	assert.Equal(t, "https://github.com/owner/repo/issues/42", issues[0].WebURL)
}

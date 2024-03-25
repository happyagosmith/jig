package clients_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/happyagosmith/jig/internal/entities"
	"github.com/happyagosmith/jig/internal/repo/clients"
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

	g, err := clients.NewGitLab(server.URL, "token")
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

	g, err := clients.NewGitLab(server.URL, "token")
	assert.NoError(t, err)
	releaseURL, err := g.GetReleaseURL(gitRepoID, version)

	assert.NoError(t, err)
	assert.Equal(t, "https://gitlab.example.com/my/repo/releases/v1.0.0", releaseURL)
}

func TestGetCommits(t *testing.T) {
	gitSrv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/api/v4/projects/1/repository/compare" {
			rw.Write([]byte(`{
            "commits": [
                {
                    "id": "1",
                    "short_id": "1",
                    "title": "Test commit",
                    "message": "This is a test commit",
					"created_at": "2021-01-01T00:00:00Z",
					"web_url": "http://gitlab.example.com/my/repo/commit/1"
                }
            ]
        }`))
		} else {
			http.Error(rw, "Not found", http.StatusNotFound)
		}
	}))
	defer gitSrv.Close()

	g, err := clients.NewGitLab(gitSrv.URL, "token")
	assert.NoError(t, err)

	id, from, to := "1", "0.0.0", "0.0.1"
	wantCreatedAt := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	wantCommit := entities.RepoRecord{
		ID:        "1",
		ShortID:   "1",
		Title:     "Test commit",
		Message:   "This is a test commit",
		CreatedAt: &wantCreatedAt,
		WebURL:    "http://gitlab.example.com/my/repo/commit/1",
		Origin:    "commit",
	}

	commits, err := g.GetCommits(id, from, to)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(commits) != 1 {
		t.Fatalf("Expected %d commits, got %d", 1, len(commits))
	}

	assert.Equal(t, wantCommit.ID, commits[0].ID)
	assert.Equal(t, wantCommit.ShortID, commits[0].ShortID)
	assert.Equal(t, wantCommit.Title, commits[0].Title)
	assert.Equal(t, wantCommit.Message, commits[0].Message)
	assert.Equal(t, wantCommit.CreatedAt, commits[0].CreatedAt)
	assert.Equal(t, wantCommit.WebURL, commits[0].WebURL)
	assert.Equal(t, wantCommit.Origin, commits[0].Origin)
}

func TestGetMergeRequests(t *testing.T) {
	gotParams := url.Values{}
	gitSrv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/api/v4/projects/1/merge_requests" {
			gotParams = req.URL.Query()
			rw.Write([]byte(`[
					{
						"id": 10,
						"iid": 1,
						"title": "this is a merge request",
						"description": "this is a merge request description",
						"sha": "commit1",
						"web_url": "http://gitlab.example.com/my/repo/merge_requests/1",
						"merged_at": "2021-01-01T00:00:00Z"
					},
					{
						"id": 12,
						"iid": 2,
						"title": "this is a merge request",
						"description": "this is a merge request description",
						"sha": "commit2",
						"web_url": "http://gitlab.example.com/my/repo/merge_requests/2",
						"merged_at": "2021-01-02T00:00:00Z"
					}
				]`))
		} else {
			http.Error(rw, "Not found", http.StatusNotFound)
		}
	}))
	defer gitSrv.Close()

	g, err := clients.NewGitLab(gitSrv.URL, "token")
	assert.NoError(t, err)

	wantCreatedAt := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	wantMr := entities.RepoRecord{
		ID:        "10",
		ShortID:   "1",
		Title:     "this is a merge request",
		Message:   "this is a merge request description",
		Origin:    "merge_request",
		CreatedAt: &wantCreatedAt,
		WebURL:    "http://gitlab.example.com/my/repo/merge_requests/1",
	}

	wantUpdatedAfter := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	wantUpdatedAfterStr := wantUpdatedAfter.Format(time.RFC3339)

	commits := []entities.RepoRecord{
		{
			ID:        "commit1",
			ShortID:   "commit1",
			Title:     "Test commit",
			Message:   "This is a test commit",
			CreatedAt: &wantUpdatedAfter,
		},
	}

	mrs, err := g.GetMergeRequests("1", "master", commits)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	assert.Equal(t, wantUpdatedAfterStr, gotParams.Get("updated_after"))
	assert.Equal(t, "master", gotParams.Get("target_branch"))
	assert.Equal(t, "merged", gotParams.Get("state"))

	if len(mrs) != 1 {
		t.Fatalf("Expected %d merge requests, got %d", 1, len(mrs))
	}

	assert.Equal(t, wantMr.ID, mrs[0].ID)
	assert.Equal(t, wantMr.ShortID, mrs[0].ShortID)
	assert.Equal(t, wantMr.Title, mrs[0].Title)
	assert.Equal(t, wantMr.Message, mrs[0].Message)
	assert.Equal(t, wantMr.CreatedAt, mrs[0].CreatedAt)
	assert.Equal(t, wantMr.WebURL, mrs[0].WebURL)
	assert.Equal(t, wantMr.Origin, mrs[0].Origin)
}

package issuetrackers_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/happyagosmith/jig/internal/entities"
	"github.com/happyagosmith/jig/internal/issuetrackers"
	"github.com/stretchr/testify/assert"
)

func TestJiraGetKnownIssues(t *testing.T) {
	tests := []struct {
		name          string
		project       string
		component     string
		expectedQuery string
	}{
		{
			name:          "test jira GetKnownIssues1",
			project:       "project",
			component:     "component",
			expectedQuery: "jql=key%3Dvalue+and+project+%3D+%22project%22+and+component+%3D+%22component%22&maxResults=1000",
		},
		{
			name:          "test jira GetKnownIssues3",
			project:       "project",
			component:     "",
			expectedQuery: "jql=key%3Dvalue+and+project+%3D+%22project%22&maxResults=1000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotRequest *http.Request
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotRequest = r
				w.WriteHeader(200)
				b, _ := os.ReadFile("testdata/jira-issues.json")
				w.Write(b)
			}))
			defer srv.Close()

			j, err := issuetrackers.NewJira(srv.URL, "jiraUsername", "jiraPassword",
				issuetrackers.WithKnownIssueJql("key=value"))
			assert.NoError(t, err, "NewJira error must be nil")

			_, err = j.GetKnownIssues(&entities.Repo{Project: tt.project, Component: tt.component})
			assert.NoError(t, err, "GetIssues error must be nil")
			assert.Equal(t, tt.expectedQuery, gotRequest.URL.RawQuery)
		})
	}
}

func TestJira(t *testing.T) {

	t.Run("test jira GetIssues with no commits", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			b, _ := os.ReadFile("testdata/jira-issues.json")
			w.Write(b)
		}))
		defer srv.Close()

		jira, err := issuetrackers.NewJira(srv.URL, "jiraUsername", "jiraPassword",
			issuetrackers.WithClosedFeatureFilter("STORY", "GOLIVE"),
			issuetrackers.WithClosedFeatureFilter("TECH TASK", "Completata"),
			issuetrackers.WithFixedBugFilter("BUG", "FIXED"),
			issuetrackers.WithFixedBugFilter("BUG", "RELEASED"),
		)
		assert.NoError(t, err, "NewJira error must be nil")

		i, err := jira.GetIssues(&entities.Repo{}, []string{})
		assert.NoError(t, err, "GetIssues error must be nil")

		assert.True(t, len(i) == 0)
	})
	t.Run("test jira GetIssues", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			b, _ := os.ReadFile("testdata/jira-issues.json")
			w.Write(b)
		}))
		defer srv.Close()

		jira, err := issuetrackers.NewJira(srv.URL, "jiraUsername", "jiraPassword",
			issuetrackers.WithClosedFeatureFilter("TECH TASK", "Completata"),
			issuetrackers.WithFixedBugFilter("BUG", "FIXED"),
			issuetrackers.WithFixedBugFilter("BUG", "RELEASED"),
		)
		assert.NoError(t, err, "NewJira error must be nil")

		i, err := jira.GetIssues(&entities.Repo{}, []string{"test"})
		assert.NoError(t, err, "GetIssues error must be nil")

		assert.True(t, i[0].Category == entities.OTHER)
		assert.True(t, i[1].Category == entities.CLOSED_FEATURE)
		assert.True(t, i[2].Category == entities.FIXED_BUG)
		assert.True(t, i[3].Category == entities.FIXED_BUG)
		assert.True(t, i[4].Category == entities.SUB_TASK)
	})

	t.Run("test jira GetIssues with subtask", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			b, _ := os.ReadFile("testdata/jira-subtask.json")
			w.Write(b)
		}))
		defer srv.Close()

		jira, err := issuetrackers.NewJira(srv.URL, "jiraUsername", "jiraPassword",
			issuetrackers.WithClosedFeatureFilter("STORY", "GOLIVE"),
			issuetrackers.WithClosedFeatureFilter("TECH TASK", "Completata"),
			issuetrackers.WithFixedBugFilter("BUG", "FIXED"),
			issuetrackers.WithFixedBugFilter("BUG", "RELEASED"),
		)
		assert.NoError(t, err, "NewJira error must be nil")

		i, err := jira.GetIssues(&entities.Repo{}, []string{"test"})
		assert.NoError(t, err, "GetIssues error must be nil")

		assert.True(t, len(i) == 2)
	})

	t.Run("test jira GetKnownIssues", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			b, _ := os.ReadFile("testdata/jira-issues.json")
			w.Write(b)
		}))
		defer srv.Close()

		j, err := issuetrackers.NewJira(srv.URL, "jiraUsername", "jiraPassword",
			issuetrackers.WithClosedFeatureFilter("STORY", "GOLIVE"),
			issuetrackers.WithKnownIssueJql("key=value"))
		assert.NoError(t, err, "NewJira error must be nil")

		issues, err := j.GetKnownIssues(&entities.Repo{Project: "TEST", Component: ""})
		assert.NoError(t, err, "GetIssues error must be nil")
		assert.Equal(t, 5, len(issues))
		assert.Equal(t, entities.CLOSED_FEATURE, issues[0].Category)
		assert.Equal(t, "AAA-0", issues[0].IssueKey)
		assert.Equal(t, "this is a story", issues[0].IssueSummary)
		assert.Equal(t, "Story", issues[0].IssueType)
		assert.Equal(t, "GOLIVE", issues[0].IssueStatus)
	})

	t.Run("test api error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(404)
		}))
		defer srv.Close()

		jira, err := issuetrackers.NewJira(srv.URL, "jiraUsername", "jiraPassword")
		assert.NoError(t, err, "NewJira error must be nil")

		_, err = jira.GetIssues(&entities.Repo{}, []string{"test"})
		assert.NotNil(t, err, "GetIssues should return error")
	})

}

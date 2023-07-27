package trackers_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/happyagosmith/jig/internal/git"
	"github.com/happyagosmith/jig/internal/model"
	"github.com/happyagosmith/jig/internal/trackers"
	"github.com/stretchr/testify/assert"
)

func TestJira(t *testing.T) {
	t.Run("test jira issue types", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			b, _ := os.ReadFile("data/jira-issues.json")
			w.Write(b)
		}))
		defer srv.Close()

		jira, err := trackers.NewJira(srv.URL, "jiraUsername", "jiraPassword",
			trackers.WithClosedFeatureFilter("STORY", "GOLIVE"),
			trackers.WithClosedFeatureFilter("TECH TASK", "Completata"),
			trackers.WithFixedBugFilter("BUG", "FIXED"),
			trackers.WithFixedBugFilter("BUG", "RELEASED"),
		)
		assert.NoError(t, err, "NewJira error must be nil")

		i, err := jira.GetIssues([]git.CommitDetail{{IssueKey: "test"}})
		assert.NoError(t, err, "GetIssues error must be nil")

		assert.True(t, i[0].Category == model.CLOSED_FEATURE)
		assert.True(t, i[1].Category == model.CLOSED_FEATURE)
		assert.True(t, i[2].Category == model.FIXED_BUG)
		assert.True(t, i[3].Category == model.FIXED_BUG)
		assert.True(t, i[4].Category == model.SUB_TASK)
	})

	t.Run("test api error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(404)
		}))
		defer srv.Close()

		jira, err := trackers.NewJira(srv.URL, "jiraUsername", "jiraPassword")
		assert.NoError(t, err, "NewJira error must be nil")

		_, err = jira.GetIssues([]git.CommitDetail{{IssueKey: "test"}})
		assert.NotNil(t, err, "GetIssues should return error")
	})

}

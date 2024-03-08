package parsers_test

import (
	"reflect"
	"testing"

	"github.com/happyagosmith/jig/internal/parsers"
)

func TestITParser_Parse(t *testing.T) {
	parser := parsers.NewIssueExtractor(
		parsers.WithIssueTracker(
			parsers.IssuePattern{IssueTracker: "jira", Pattern: `j_(.+)`}),
		parsers.WithIssueTracker(
			parsers.IssuePattern{IssueTracker: "jira", Pattern: `JIRA-\d+`}),
		parsers.WithIssueTracker(
			parsers.IssuePattern{IssueTracker: "git", Pattern: "#([A-Z0-9]+)"}))

	tests := []struct {
		name     string
		sToParse string
		want     *parsers.IssueDetail
	}{
		{
			name:     "Test with j_JIRA-123",
			sToParse: "j_JIRA-123",
			want:     &parsers.IssueDetail{Key: "JIRA-123", IssueTracker: "JIRA"},
		},
		{
			name:     "Test with JIRA-123",
			sToParse: "JIRA-123",
			want:     &parsers.IssueDetail{Key: "JIRA-123", IssueTracker: "JIRA"},
		},
		{
			name:     "Test with #123",
			sToParse: "#123",
			want:     &parsers.IssueDetail{Key: "123", IssueTracker: "GIT"},
		},
		{
			name:     "Test with no JIRA issue",
			sToParse: "p_NOTRACKER-123",
			want:     &parsers.IssueDetail{Key: "p_NOTRACKER-123", IssueTracker: "NONE"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parser.Parse(tt.sToParse); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ITParser.Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}

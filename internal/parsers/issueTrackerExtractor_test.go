package parsers_test

import (
	"reflect"
	"testing"

	"github.com/happyagosmith/jig/internal/parsers"
)

func TestITParser_Parse(t *testing.T) {
	const pattern = `(^j_(?P<jira_1>.*)$)|(?P<jira_2>^[^\_]+$)`
	parser := parsers.NewIT(pattern)

	tests := []struct {
		name     string
		sToParse string
		want     *parsers.IssueDetail
	}{
		{
			name:     "Test with JIRA-123",
			sToParse: "j_JIRA-123",
			want:     &parsers.IssueDetail{Key: "JIRA-123", IssueTracker: parsers.JIRA},
		},
		{
			name:     "Test with JIRA123",
			sToParse: "JIRA-123",
			want:     &parsers.IssueDetail{Key: "JIRA-123", IssueTracker: parsers.JIRA},
		},
		{
			name:     "Test with no JIRA issue",
			sToParse: "p_NOJIRA-123",
			want:     nil,
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

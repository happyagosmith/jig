package parsers_test

import (
	"strings"
	"testing"

	"github.com/happyagosmith/jig/internal/parsers"
	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	verbGroups := map[parsers.Verb][]string{
		"close":     {"Close", "Closes", "Closed", "Closing", "close", "closes", "closed", "closing"},
		"fix":       {"Fix", "Fixes", "Fixed", "Fixing", "fix", "fixes", "fixed", "fixing"},
		"resolve":   {"Resolve", "Resolves", "Resolved", "Resolving", "resolve", "resolves", "resolved", "resolving"},
		"implement": {"Implements", "Implemented", "Implementing", "implement", "implements", "implemented", "implementing"},
	}

	tests := []struct {
		name    string
		input   string
		wantKey []string
		wantErr bool
	}{
		{
			name:    "verb single issue",
			input:   "verb #123",
			wantKey: []string{"#123"},
			wantErr: false,
		},
		{
			name:    "verb multiple issues separated by coma",
			input:   "verb #123, JIRA-456, #789",
			wantKey: []string{"#123", "JIRA-456", "#789"},
			wantErr: false,
		},
		{
			name:    "verb multiple issues separated by and",
			input:   "verb #123 and JIRA-456",
			wantKey: []string{"#123", "JIRA-456"},
			wantErr: false,
		},
		{
			name:    "verb multiple issues separated by coma and and",
			input:   "verb #123, #456 and JIRA-789",
			wantKey: []string{"#123", "#456", "JIRA-789"},
			wantErr: false,
		},
		{
			name:    "verb multiple issues",
			input:   "verb #123, verb JIRA-456 and verb #789",
			wantKey: []string{"#123", "JIRA-456", "#789"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		for verb, group := range verbGroups {
			for _, v := range group {
				name := strings.Replace(tt.name, "verb", v, 1)
				t.Run(name, func(t *testing.T) {
					str := strings.ReplaceAll(tt.input, "verb", v)
					pcp := parsers.NewClosingPattern(
						parsers.WithIssuePattern("#([A-Z0-9]+)"),
						parsers.WithIssuePattern("JIRA-[0-9]+"))
					got, err := pcp.Parse(str)
					assert.NoError(t, err)
					assert.Equal(t, len(tt.wantKey), len(got))
					for i := range got {
						assert.Equal(t, tt.wantKey[i], got[i].Key)
						assert.Equal(t, verb, got[i].Verb)
					}
				})
			}
		}
	}
}

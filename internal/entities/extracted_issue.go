package entities

import "fmt"

type ExtractedIssue struct {
	IssueTracker  string        `yaml:"issueTracker"`
	IssueKey      string        `yaml:"issueKey,omitempty"`
	IssueSummary  string        `yaml:"issueSummary,omitempty"`
	IssueCategory IssueCategory `yaml:"issueCategory"`
	Issue         `yaml:"issueDetail,omitempty"`
	ParsedCommit  `yaml:"commitDetail,omitempty"`
}

func (i ExtractedIssue) String() string {
	return fmt.Sprintf("key %s, issue type %s", i.Issue.IssueKey, i.Issue.Category)
}

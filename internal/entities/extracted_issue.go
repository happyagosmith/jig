package entities

import "fmt"

type ExtractedIssue struct {
	IssueTracker string `yaml:"issueTracker"`
	Issue        `yaml:",inline"`
	ParsedCommit `yaml:"commitDetail,omitempty"`
}

func (i ExtractedIssue) String() string {
	return fmt.Sprintf("key %s, issue type %s", i.Issue.IssueKey, i.Issue.Category)
}

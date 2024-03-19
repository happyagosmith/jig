package entities

import "fmt"

type ExtractedIssue struct {
	IssueTracker     string        `yaml:"issueTracker"`
	IssueKey         string        `yaml:"issueKey,omitempty"`
	IssueSummary     string        `yaml:"issueSummary,omitempty"`
	IssueCategory    IssueCategory `yaml:"issueCategory"`
	Issue            `yaml:"issueDetail,omitempty"`
	ParsedRepoRecord `yaml:"repoDetail,omitempty"`
}

func (i ExtractedIssue) String() string {
	return fmt.Sprintf("%s issue %s", i.IssueTracker, i.Issue.String())
}

package trackers

import (
	"fmt"
	"strings"
)

type CategoryType int

const (
	CLOSED_FEATURE CategoryType = iota
	FIXED_BUG
	SUB_TASK
	OTHER
)

func (ct CategoryType) String() string {
	return []string{"CLOSED_FEATURE", "FIXED_BUG", "SUB_TASK", "OTHER"}[ct]
}

func (ct CategoryType) MarshalYAML() (interface{}, error) {
	return ct.String(), nil
}

func (ct *CategoryType) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}

	switch strings.ToLower(s) {
	case "closed_feature":
		*ct = CLOSED_FEATURE
	case "fixed_bug":
		*ct = FIXED_BUG
	case "sub_task":
		*ct = SUB_TASK
	case "other":
		*ct = OTHER
	default:
		return fmt.Errorf("invalid CategoryType %q", s)
	}

	return nil
}

type IssueDetail struct {
	Category     CategoryType `yaml:"category"`
	IssueKey     string       `yaml:"issueKey,omitempty"`
	IssueSummary string       `yaml:"issueSummary,omitempty"`
	IssueType    string       `yaml:"issueType,omitempty"`
	IssueStatus  string       `yaml:"issueStatus,omitempty"`
}

func (i IssueDetail) String() string {
	return fmt.Sprintf("key %s, issue type %s", i.IssueKey, i.Category)
}

package entities

import (
	"fmt"
	"strings"
)

type CommitCategory int

const (
	UNKNOWN CommitCategory = iota
	FEATURE
	BUG_FIX
)

func (i CommitCategory) String() string {
	return []string{"UNKNOWN", "FEATURE", "BUG_FIX"}[i]
}

func (s CommitCategory) MarshalYAML() (interface{}, error) {
	return s.String(), nil
}

func (cct *CommitCategory) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}

	switch strings.ToLower(s) {
	case "unknown":
		*cct = UNKNOWN
	case "feature":
		*cct = FEATURE
	case "bug_fix":
		*cct = BUG_FIX
	default:
		return fmt.Errorf("invalid CCType %q", s)
	}

	return nil
}

type ParsedRepoRecord struct {
	RepoRecord         `yaml:",inline"`
	ParsedSummary      string         `yaml:"parsedSummary,omitempty"`
	ParsedCategory     CommitCategory `yaml:"parsedCategory,omitempty"`
	ParsedKey          string         `yaml:"parsedKey,omitempty"`
	ParsedIssueTracker string         `yaml:"parsedIssueTracker"`
	Parser             string         `yaml:"parser,omitempty"`
	ParsedType         string         `yaml:"parsedType,omitempty"`
	IsBreakingChange   bool           `yaml:"isBreakingChange,omitempty"`
}

func (c ParsedRepoRecord) String() string {
	return fmt.Sprintf("%s issue %s (%s) with %s parser on %s", c.ParsedIssueTracker, c.ParsedKey, c.ParsedType, c.Parser, c.RepoRecord.String())
}

func (c ParsedRepoRecord) ShortString() string {
	return fmt.Sprintf("%s (%s -> %s)", c.ParsedKey, c.ParsedType, c.ParsedCategory)
}

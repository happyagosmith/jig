package parsers

import (
	"fmt"
	"regexp"
	"strings"
)

type ITParser struct {
	re  regexp.Regexp
	gni map[string][]int
}

type IssueTrackerType int

const (
	NONE IssueTrackerType = iota
	JIRA
)

func (itt IssueTrackerType) String() string {
	return []string{"NONE", "JIRA"}[itt]
}

func (itt IssueTrackerType) MarshalYAML() (interface{}, error) {
	return itt.String(), nil
}

func (itt *IssueTrackerType) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}

	switch strings.ToLower(s) {
	case "none":
		*itt = NONE
	case "jira":
		*itt = JIRA
	default:
		return fmt.Errorf("invalid IssueTrackerType %q", s)
	}

	return nil
}

type IssueDetail struct {
	Key          string
	IssueTracker IssueTrackerType
}

func NewIT(pattern string) ITParser {
	re := regexp.MustCompile(pattern)
	gn := re.SubexpNames()
	gnidx := map[string][]int{}

	for i, n := range gn {
		if n == "" {
			continue
		}
		it := strings.Split(n, "_")[0]
		// for now we only support jira
		if it != "jira" {
			continue
		}
		if _, ok := gnidx[it]; !ok {
			gnidx[it] = []int{i}
			continue
		}
		gnidx[it] = append(gnidx[it], i)
	}

	return ITParser{re: *re, gni: gnidx}
}

func (p ITParser) Parse(sToParse string) *IssueDetail {
	if sToParse == "" {
		return nil
	}
	gitIssues := p.re.FindAllStringSubmatch(sToParse, -1)
	if len(gitIssues) == 0 {
		return nil
	}

	for _, idx := range p.gni["jira"] {
		if gitIssues[0][idx] != "" {
			return &IssueDetail{Key: gitIssues[0][idx], IssueTracker: JIRA}
		}
	}

	return nil
}

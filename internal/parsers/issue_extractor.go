package parsers

import (
	"regexp"
	"strings"
)

const UNKNOWN_ISSUE_TRACKER = "NONE"

type IssuePattern struct {
	IssueTracker string
	Pattern      string
}

type IssueExtractor struct {
	re  []regexp.Regexp
	ips []IssuePattern
}

type IssueDetail struct {
	Key          string
	IssueTracker string
}

type IssueExtractorOpt func(*IssueExtractor)

func WithIssueTracker(it IssuePattern) IssueExtractorOpt {
	return func(p *IssueExtractor) {
		if it.IssueTracker != "" && it.Pattern != "" {
			re := regexp.MustCompile(it.Pattern)

			p.re = append(p.re, *re)
			p.ips = append(p.ips, IssuePattern{IssueTracker: strings.ToUpper(it.IssueTracker), Pattern: it.Pattern})
		}
	}
}

func NewIssueExtractor(opts ...IssueExtractorOpt) IssueExtractor {
	p := IssueExtractor{}
	for _, o := range opts {
		o(&p)
	}

	return p
}

func (p IssueExtractor) Parse(sToParse string) *IssueDetail {
	if sToParse == "" {
		return &IssueDetail{Key: sToParse, IssueTracker: UNKNOWN_ISSUE_TRACKER}
	}

	for i, re := range p.re {
		matches := re.FindAllStringSubmatch(sToParse, -1)
		if len(matches) == 0 {
			continue
		}

		idx := len(matches[0]) - 1
		key := matches[0][idx]
		if key != "" {
			return &IssueDetail{Key: key, IssueTracker: p.ips[i].IssueTracker}
		}
	}

	return &IssueDetail{Key: sToParse, IssueTracker: UNKNOWN_ISSUE_TRACKER}
}

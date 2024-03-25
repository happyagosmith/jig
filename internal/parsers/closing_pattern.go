package parsers

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/happyagosmith/jig/internal/entities"
)

type Verb string

const (
	Close     Verb = "close"
	Implement Verb = "implement"
	Fix       Verb = "fix"
	Resolve   Verb = "resolve"
)

func (v Verb) String() string {
	return string(v)
}

type CPIssue struct {
	Key      string
	Category entities.CommitCategory
	Verb     Verb
}

type ClosingPatternParser struct {
	issueRefs []string
	re        *regexp.Regexp
}

type ClosingPatternOpt func(*ClosingPatternParser)

func WithIssuePattern(pattern string) ClosingPatternOpt {
	return func(p *ClosingPatternParser) {
		if pattern != "" {
			p.issueRefs = append(p.issueRefs, pattern)
		}
	}
}

func NewClosingPattern(opts ...ClosingPatternOpt) ClosingPatternParser {
	p := ClosingPatternParser{}
	for _, o := range opts {
		o(&p)
	}

	issueRefs := strings.Join(p.issueRefs, "|")
	pattern := fmt.Sprintf(`(?m)\b(?:[Cc]los(?:e[sd]?|ing)|\b[Ff]ix(?:e[sd]|ing)?|\b[Rr]esolv(?:e[sd]?|ing)|\b[Ii]mplement(?:s|ed|ing)?)(:?) +(?:(?:issues? +)?(?:%s)(?:(?: *,? +and +| *,? *)?)|([A-Z][A-Z0-9_]+-\d+))+`, issueRefs)
	re := regexp.MustCompile(pattern)
	p.re = re

	return p
}

func (p ClosingPatternParser) Parse(s string) ([]CPIssue, error) {
	matches := p.re.FindAllString(s, -1)

	if matches == nil {
		return nil, nil
	}

	result := make([]CPIssue, 0)
	for _, match := range matches {
		v := strings.Split(match, " ")[0]

		keys := strings.TrimLeft(match, v+" ")
		verb := extractVerb(v)

		for _, key := range split(keys) {
			category := entities.UNKNOWN
			if verb == Close || verb == Implement {
				category = entities.FEATURE
			} else if verb == Fix || verb == Resolve {
				category = entities.BUG_FIX
			}
			result = append(result, CPIssue{
				Verb:     verb,
				Key:      key,
				Category: category,
			})
		}
	}

	return result, nil
}

func split(s string) []string {
	result := make([]string, 0)
	for _, v := range strings.Split(s, ",") {
		for _, v2 := range strings.Split(v, "and") {
			str := strings.TrimSpace(v2)
			if str == "" {
				continue
			}
			result = append(result, str)
		}
	}
	return result
}

func extractVerb(s string) Verb {
	v := strings.ToLower(s)
	switch {
	case strings.HasPrefix(v, "clos"):
		return Close
	case strings.HasPrefix(v, "fix"):
		return Fix
	case strings.HasPrefix(v, "resolv"):
		return Resolve
	case strings.HasPrefix(v, "implement"):
		return Implement
	default:
		return ""
	}
}

package parsers

import (
	"fmt"
	"regexp"
	"strings"
)

type CCParser struct {
	re  regexp.Regexp
	gni map[string]int
}

type CCType int

const (
	UNKNOWN CCType = iota
	FEATURE
	BUG_FIX
)

func (i CCType) String() string {
	return []string{"UNKNOWN", "FEATURE", "BUG_FIX"}[i]
}

func (s CCType) MarshalYAML() (interface{}, error) {
	return s.String(), nil
}

func (cct *CCType) UnmarshalYAML(unmarshal func(interface{}) error) error {
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

type ConventionalCommit struct {
	Type       string
	Category   CCType
	Scope      string
	IsBreaking bool
	Subject    string
}

const ccPattern = `^(?P<type>[^\(\:]*)(\((?P<scope>[^\)]+)\))?(?P<breaking>!)?: (?P<subject>.*)?`

func NewConventionalCommit() CCParser {
	re := regexp.MustCompile(ccPattern)
	gn := re.SubexpNames()
	gnidx := map[string]int{}

	for i, n := range gn {
		if n != "" {
			gnidx[n] = i
		}
	}

	return CCParser{re: *re, gni: gnidx}
}

func (p CCParser) Parse(commit string) *ConventionalCommit {
	cc := p.re.FindAllStringSubmatch(commit, -1)

	if len(cc) == 0 {
		return nil
	}

	cct := UNKNOWN
	t := cc[0][p.gni["type"]]
	if t == "feat" {
		cct = FEATURE
	}

	if t == "fix" {
		cct = BUG_FIX
	}

	isBreaking := false
	if cc[0][p.gni["breaking"]] == "!" {
		isBreaking = true
	}

	if strings.Contains(commit, "BREAKING CHANGE: ") {
		isBreaking = true
	}

	return &ConventionalCommit{
		Type:       t,
		Category:   cct,
		Scope:      cc[0][p.gni["scope"]],
		Subject:    cc[0][p.gni["subject"]],
		IsBreaking: isBreaking,
	}
}

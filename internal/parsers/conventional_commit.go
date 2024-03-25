package parsers

import (
	"regexp"
	"strings"

	"github.com/happyagosmith/jig/internal/entities"
)

type CCParser struct {
	re  regexp.Regexp
	gni map[string]int
}

type ConventionalCommit struct {
	Type       string
	Category   entities.CommitCategory
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

	cct := entities.UNKNOWN
	t := cc[0][p.gni["type"]]
	if t == "feat" {
		cct = entities.FEATURE
	}

	if t == "fix" {
		cct = entities.BUG_FIX
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

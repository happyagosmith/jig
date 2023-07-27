package parsers

import (
	"regexp"
	"strings"
)

const customPattern = `\[(?P<scope>[^\]]*)\](?P<subject>.*)`

type CustomParser struct {
	re      regexp.Regexp
	gni     map[string]int
	pattern string
}

type ParserOpt func(*CustomParser)

func WithPattern(pattern string) ParserOpt {
	return func(p *CustomParser) {
		if pattern != "" {
			p.pattern = pattern
		}
	}
}

func NewCustom(opts ...ParserOpt) CustomParser {
	p := CustomParser{}
	p.pattern = customPattern
	for _, o := range opts {
		o(&p)
	}
	re := regexp.MustCompile(p.pattern)
	gn := re.SubexpNames()
	gnidx := map[string]int{}

	for i, n := range gn {
		if n != "" {
			gnidx[n] = i
		}
	}

	p.re = *re
	p.gni = gnidx

	return p
}

func (p CustomParser) Parse(commit string) *ConventionalCommit {
	cc := p.re.FindAllStringSubmatch(commit, -1)

	if len(cc) == 0 {
		return nil
	}

	return &ConventionalCommit{
		Scope:   cc[0][p.gni["scope"]],
		Subject: strings.TrimLeft(cc[0][p.gni["subject"]], " "),
	}
}

package parsers

import (
	"fmt"

	"github.com/happyagosmith/jig/internal/entities"
)

type Repo struct {
	conventionalParser CCParser
	itParser           IssueExtractor
	customParser       *CustomParser
	closingPattern     ClosingPatternParser
	keepCCWithoutScope bool
}

type RepoParserOpt func(*Repo)

func WithKeepCCWithoutScope(v bool) RepoParserOpt {
	return func(j *Repo) {
		j.keepCCWithoutScope = v
	}
}

func WithCustomPattern(v string) RepoParserOpt {
	return func(j *Repo) {
		if v != "" {
			c := NewCustomCommit(WithPattern(v))
			j.customParser = &c
		}
	}
}

func New(issuePatterns []IssuePattern, opts ...RepoParserOpt) (Repo, error) {

	itOpts := make([]IssueExtractorOpt, 0, len(issuePatterns))
	cpOpts := make([]ClosingPatternOpt, 0, len(issuePatterns))

	for it := range issuePatterns {
		itOpts = append(itOpts, WithIssueTracker(issuePatterns[it]))
		cpOpts = append(cpOpts, WithIssuePattern(issuePatterns[it].Pattern))
	}

	g := Repo{
		conventionalParser: NewConventionalCommit(),
		itParser:           NewIssueExtractor(itOpts...),
		closingPattern:     NewClosingPattern(cpOpts...),
	}
	for _, o := range opts {
		o(&g)
	}

	return g, nil
}

func (g Repo) Parse(commits []entities.Commit) ([]entities.ParsedCommit, error) {
	found := map[string]bool{}
	var cds []entities.ParsedCommit

	for _, commit := range commits {
		fmt.Printf("processing commit %s \n", commit.ShortID)
		cd := entities.ParsedCommit{}
		cd.Parser = "conventionalParser"

		cc := g.conventionalParser.Parse(commit.Message)
		if cc == nil && g.customParser != nil {
			cc = g.customParser.Parse(commit.Message)
			cd.Parser = "customParser"
		}
		if cc != nil && (cc.Scope != "" || (g.keepCCWithoutScope && cc.Category != entities.UNKNOWN && cc.Subject != "")) {
			cd.ParsedKey = cc.Scope
			cd.ParsedCategory = cc.Category
			cd.Summary = cc.Subject
			cd.Message = commit.Message
			cd.CommitID = commit.ID
			cd.IsBreakingChange = cc.IsBreaking
			cd.ParsedType = cc.Type
			issueDetails := g.itParser.Parse(cc.Scope)
			cd.ParsedKey = issueDetails.Key
			cd.ParsedIssueTracker = issueDetails.IssueTracker
			if cd.ParsedKey == "" {
				cds = append(cds, cd)
				fmt.Printf("added conventional commit without issueKey \"%s\" \n", cd.Message)
				continue
			}
			if !found[cd.ParsedKey] {
				found[cd.ParsedKey] = true
				cds = append(cds, cd)
				fmt.Printf("extracted key %s \n", cd.ParsedKey)
			}
		}

		cpc, _ := g.closingPattern.Parse(commit.Message)
		for _, c := range cpc {
			issueDetails := g.itParser.Parse(c.Key)
			cds = append(cds, entities.ParsedCommit{
				ParsedKey:          issueDetails.Key,
				ParsedIssueTracker: issueDetails.IssueTracker,
				ParsedCategory:     c.Category,
				Summary:            commit.Title,
				Message:            commit.Message,
				CommitID:           commit.ID,
				IsBreakingChange:   false,
				Parser:             "closingPattern",
				ParsedType:         c.Verb.String(),
			})
		}
	}

	return cds, nil
}

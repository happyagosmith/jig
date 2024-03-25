package repo

import (
	"fmt"

	"github.com/happyagosmith/jig/internal/entities"
	"github.com/happyagosmith/jig/internal/parsers"
)

type Repo struct {
	conventionalParser    parsers.CCParser
	itParser              parsers.IssueExtractor
	customParser          *parsers.CustomParser
	closingPattern        parsers.ClosingPatternParser
	keepCCWithoutScope    bool
	repoClient            entities.RepoClient
	defaultMRTargetBranch string
}

type RepoParserOpt func(*Repo)

func WithKeepCCWithoutScope(v bool) RepoParserOpt {
	return func(r *Repo) {
		r.keepCCWithoutScope = v
	}
}

func WithCustomPattern(v string) RepoParserOpt {
	return func(r *Repo) {
		if v != "" {
			c := parsers.NewCustomCommit(parsers.WithPattern(v))
			r.customParser = &c
		}
	}
}

func WithDefaultMRBranch(v string) RepoParserOpt {
	return func(r *Repo) {
		if v != "" {
			r.defaultMRTargetBranch = v
		}
	}
}

func New(client entities.RepoClient, issuePatterns []parsers.IssuePattern, opts ...RepoParserOpt) (Repo, error) {

	itOpts := make([]parsers.IssueExtractorOpt, 0, len(issuePatterns))
	cpOpts := make([]parsers.ClosingPatternOpt, 0, len(issuePatterns))

	for it := range issuePatterns {
		itOpts = append(itOpts, parsers.WithIssueTracker(issuePatterns[it]))
		cpOpts = append(cpOpts, parsers.WithIssuePattern(issuePatterns[it].Pattern))
	}

	g := Repo{
		conventionalParser: parsers.NewConventionalCommit(),
		itParser:           parsers.NewIssueExtractor(itOpts...),
		closingPattern:     parsers.NewClosingPattern(cpOpts...),
		repoClient:         client,
	}
	for _, o := range opts {
		o(&g)
	}

	return g, nil
}

func (r Repo) GetParsedRecords(id, from, to, mrTargetBranch string) ([]entities.ParsedRepoRecord, error) {
	commits, err := r.repoClient.GetCommits(id, from, to)
	if err != nil {
		return nil, err
	}

	pcommits, err := r.parse(commits)
	if err != nil {
		return nil, err
	}

	targetBranch := mrTargetBranch
	if targetBranch == "" {
		targetBranch = r.defaultMRTargetBranch
	}
	if targetBranch == "" {
		return pcommits, nil
	}
	mr, err := r.repoClient.GetMergeRequests(id, targetBranch, commits)
	if err != nil {
		return nil, err
	}
	pmr, err := r.parse(mr)
	if err != nil {
		return nil, err
	}

	pcommits = append(pcommits, pmr...)

	return pcommits, nil
}

func (r Repo) GetReleaseURL(id, tag string) (string, error) {
	return r.repoClient.GetReleaseURL(id, tag)
}

func (r Repo) GetRepoURL(id string) (string, error) {
	return r.repoClient.GetRepoURL(id)
}

func (r Repo) parse(commits []entities.RepoRecord) ([]entities.ParsedRepoRecord, error) {
	found := map[string]bool{}
	var cds []entities.ParsedRepoRecord

	for _, commit := range commits {
		fmt.Printf("parsing %s \n", commit.String())
		cd := entities.ParsedRepoRecord{RepoRecord: commit}
		cd.Parser = "conventionalParser"

		cc := r.conventionalParser.Parse(commit.Title)
		if cc == nil && r.customParser != nil {
			cc = r.customParser.Parse(commit.Title)
			cd.Parser = "customParser"
		}
		if cc != nil && (cc.Scope != "" || (r.keepCCWithoutScope && cc.Category != entities.UNKNOWN && cc.Subject != "")) {
			cd.ParsedKey = cc.Scope
			cd.ParsedCategory = cc.Category
			cd.ParsedSummary = cc.Subject
			cd.IsBreakingChange = cc.IsBreaking
			cd.ParsedType = cc.Type
			issueDetails := r.itParser.Parse(cc.Scope)
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
				fmt.Printf("extracted %s \n", cd.String())
			}
		}

		cpc, _ := r.closingPattern.Parse(commit.Message)
		for _, c := range cpc {
			issueDetails := r.itParser.Parse(c.Key)
			cd := entities.ParsedRepoRecord{
				RepoRecord:         commit,
				ParsedKey:          issueDetails.Key,
				ParsedIssueTracker: issueDetails.IssueTracker,
				ParsedCategory:     c.Category,
				ParsedSummary:      "",
				IsBreakingChange:   false,
				Parser:             "closingPattern",
				ParsedType:         c.Verb.String(),
			}
			cds = append(cds, cd)
			fmt.Printf("extracted %s \n", cd.String())
		}
	}

	return cds, nil
}

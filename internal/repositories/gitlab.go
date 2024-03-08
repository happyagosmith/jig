package repositories

import (
	"fmt"
	"log"

	"github.com/happyagosmith/jig/internal/parsers"
	"github.com/xanzy/go-gitlab"
)

type CommitDetail struct {
	Summary            string         `yaml:"summary,omitempty"`
	Message            string         `yaml:"message,omitempty"`
	CommitID           string         `yaml:"commitID,omitempty"`
	ParsedCategory     parsers.CCType `yaml:"parsedCategory,omitempty"`
	ParsedKey          string         `yaml:"parsedKey,omitempty"`
	ParsedIssueTracker string         `yaml:"parsedIssueTracker"`
	Parser             string         `yaml:"parser,omitempty"`
	ParsedType         string         `yaml:"parsedType,omitempty"`
	IsBreakingChange   bool           `yaml:"isBreakingChange,omitempty"`
}

func (c CommitDetail) String() string {
	return fmt.Sprintf("key %s, issue type %s", c.ParsedKey, c.ParsedCategory)
}

type Git struct {
	c                  *gitlab.Client
	conventionalParser parsers.CCParser
	itParser           parsers.IssueExtractor
	customParser       *parsers.CustomParser
	closingPattern     parsers.ClosingPatternParser
	keepCCWithoutScope bool
}

type GitOpt func(*Git)

func WithKeepCCWithoutScope(v bool) GitOpt {
	return func(j *Git) {
		j.keepCCWithoutScope = v
	}
}

func WithCustomPattern(v string) GitOpt {
	return func(j *Git) {
		if v != "" {
			c := parsers.NewCustomCommit(parsers.WithPattern(v))
			j.customParser = &c
		}
	}
}

func NewClient(URL, token string, issuePatterns []parsers.IssuePattern, opts ...GitOpt) (Git, error) {
	c, err := gitlab.NewClient(token,
		gitlab.WithBaseURL(fmt.Sprintf("%s/api/v4/", URL)))
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	itOpts := make([]parsers.IssueExtractorOpt, 0, len(issuePatterns))
	cpOpts := make([]parsers.ClosingPatternOpt, 0, len(issuePatterns))

	for it := range issuePatterns {
		itOpts = append(itOpts, parsers.WithIssueTracker(issuePatterns[it]))
		cpOpts = append(cpOpts, parsers.WithIssuePattern(issuePatterns[it].Pattern))
	}

	g := Git{
		c:                  c,
		conventionalParser: parsers.NewConventionalCommit(),
		itParser:           parsers.NewIssueExtractor(itOpts...),
		closingPattern:     parsers.NewClosingPattern(cpOpts...),
	}
	for _, o := range opts {
		o(&g)
	}

	return g, nil
}

func (g Git) ExtractCommits(id, from, to string) ([]CommitDetail, error) {
	var cds []CommitDetail
	opt := &gitlab.CompareOptions{From: &from, To: &to}

	c, _, err := g.c.Repositories.Compare(id, opt)
	if err != nil {
		return nil, err
	}

	found := map[string]bool{}

	for _, commit := range c.Commits {
		fmt.Printf("processing commit %s \n", commit.ShortID)
		cd := CommitDetail{}
		cd.Parser = "conventionalParser"

		cc := g.conventionalParser.Parse(commit.Message)
		if cc == nil && g.customParser != nil {
			cc = g.customParser.Parse(commit.Message)
			cd.Parser = "customParser"
		}
		if cc != nil && (cc.Scope != "" || (g.keepCCWithoutScope && cc.Category != parsers.UNKNOWN && cc.Subject != "")) {
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
			cds = append(cds, CommitDetail{
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

type ProjectResponse struct {
	WebURL string `json:"web_url"`
}

type ReleaseResponse struct {
	Links struct {
		Self string `json:"self"`
	} `json:"_links"`
}

func (g Git) GetRepoURL(gitRepoID string) (string, error) {
	p, _, err := g.c.Projects.GetProject(gitRepoID, nil)
	if err != nil {
		return "", err
	}

	repoURL := p.WebURL

	return repoURL, nil
}

func (g Git) GetReleaseURL(gitRepoID, version string) (string, error) {
	r, _, err := g.c.Releases.GetRelease(gitRepoID, version, nil)
	if err != nil {
		return "", err
	}
	releaseURL := r.Links.Self

	return releaseURL, nil
}

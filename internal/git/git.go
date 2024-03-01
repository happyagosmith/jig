package git

import (
	"fmt"
	"log"

	"github.com/happyagosmith/jig/internal/parsers"
	"github.com/xanzy/go-gitlab"
)

type CommitDetail struct {
	Category     parsers.CCType           `yaml:"category,omitempty"`
	IssueKey     string                   `yaml:"issueKey,omitempty"`
	IsBreaking   bool                     `yaml:"isBreakingChange,omitempty"`
	Summary      string                   `yaml:"summary,omitempty"`
	Message      string                   `yaml:"message,omitempty"`
	Commit       string                   `yaml:"commit,omitempty"`
	IssueTracker parsers.IssueTrackerType `yaml:"issueTrackerType"`
}

func (c CommitDetail) String() string {
	return fmt.Sprintf("key %s, issue type %s", c.IssueKey, c.Category)
}

type Git struct {
	c                  *gitlab.Client
	conventionalParser parsers.CCParser
	itParser           parsers.ITParser
	customParser       *parsers.CustomParser
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
			c := parsers.NewCustom(parsers.WithPattern(v))
			j.customParser = &c
		}
	}
}

func NewClient(URL, token, issuePattern string, opts ...GitOpt) (Git, error) {
	c, err := gitlab.NewClient(token,
		gitlab.WithBaseURL(fmt.Sprintf("%s/api/v4/", URL)))
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	g := Git{
		c:                  c,
		conventionalParser: parsers.NewCC(),
		itParser:           parsers.NewIT(issuePattern),
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
		cd := CommitDetail{IssueTracker: parsers.NONE}
		cc := g.conventionalParser.Parse(commit.Message)
		if cc == nil && g.customParser != nil {
			cc = g.customParser.Parse(commit.Message)
		}
		if cc != nil && (cc.Scope != "" || (g.keepCCWithoutScope && cc.Type != parsers.UNKNOWN && cc.Subject != "")) {
			cd.IssueKey = cc.Scope
			cd.Category = cc.Type
			cd.Summary = cc.Subject
			cd.Message = commit.Message
			cd.Commit = commit.ID
			cd.IsBreaking = cc.IsBreaking
			issueDetails := g.itParser.Parse(cc.Scope)
			if issueDetails != nil && issueDetails.Key != "" {
				cd.IssueKey = issueDetails.Key
				cd.IssueTracker = issueDetails.IssueTracker
			}
			if cd.IssueKey == "" {
				cds = append(cds, cd)
				fmt.Printf("added conventional commit without issueKey \"%s\" \n", cd.Message)
				continue
			}
			if !found[cd.IssueKey] {
				found[cd.IssueKey] = true
				cds = append(cds, cd)
				fmt.Printf("extracted key %s \n", cd.IssueKey)
			}
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

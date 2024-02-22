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
	customParser       parsers.CustomParser
}

func NewClient(URL, token, issuePattern, customPattern string) (Git, error) {
	c, err := gitlab.NewClient(token,
		gitlab.WithBaseURL(fmt.Sprintf("%s/api/v4/", URL)))
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	return Git{c: c, conventionalParser: parsers.NewCC(), itParser: parsers.NewIT(issuePattern), customParser: parsers.NewCustom(parsers.WithPattern(customPattern))}, nil
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
		if cc == nil {
			cc = g.customParser.Parse(commit.Message)
		}
		if cc != nil && cc.Scope != "" {
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
			if !found[cd.IssueKey] {
				found[cd.IssueKey] = true
				cds = append(cds, cd)
				fmt.Printf("extracted key %s \n", cd.IssueKey)
			}
		}
	}

	return cds, nil
}

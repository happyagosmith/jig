package clients

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/happyagosmith/jig/internal/entities"
	"github.com/xanzy/go-gitlab"
)

type Git struct {
	c                     *gitlab.Client
	issueLabelsForFeature []string
	issueLabelsForBug     []string
}

func NewGitLab(URL, token string) (Git, error) {
	c, err := gitlab.NewClient(token,
		gitlab.WithBaseURL(fmt.Sprintf("%s/api/v4/", URL)))
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	g := Git{
		c:                     c,
		issueLabelsForFeature: []string{"feature"},
		issueLabelsForBug:     []string{"bug"},
	}

	return g, nil
}

func (g Git) GetMergeRequests(id, targetBranch string, commits []entities.RepoRecord) ([]entities.RepoRecord, error) {
	if len(commits) == 0 {
		return nil, nil
	}

	lookForCommit := map[string]bool{}
	for _, c := range commits {
		lookForCommit[c.ID] = true
	}
	opts := &gitlab.ListProjectMergeRequestsOptions{
		UpdatedAfter: commits[0].CreatedAt,
		State:        gitlab.String("merged"),
		TargetBranch: gitlab.String(targetBranch),
	}
	mrs, _, err := g.c.MergeRequests.ListProjectMergeRequests(id, opts)

	cs := make([]entities.RepoRecord, 0, len(mrs))
	for _, mr := range mrs {
		if !lookForCommit[mr.SHA] {
			continue
		}
		cs = append(cs, entities.RepoRecord{
			ID:        strconv.Itoa(mr.ID),
			ShortID:   strconv.Itoa(mr.IID),
			Title:     mr.Title,
			Message:   mr.Description,
			CreatedAt: mr.MergedAt,
			WebURL:    mr.WebURL,
			Origin:    "merge_request",
		})
	}

	return cs, err
}

func (g Git) GetCommits(id, from, to string) ([]entities.RepoRecord, error) {
	opt := &gitlab.CompareOptions{From: &from, To: &to}

	c, _, err := g.c.Repositories.Compare(id, opt)
	if err != nil {
		return nil, err
	}

	var commits []entities.RepoRecord
	for _, commit := range c.Commits {
		commits = append(commits, entities.RepoRecord{
			ID:        commit.ID,
			ShortID:   commit.ShortID,
			Title:     commit.Title,
			Message:   commit.Message,
			CreatedAt: commit.CreatedAt,
			WebURL:    commit.WebURL,
			Origin:    "commit",
		})
	}

	return commits, nil

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

func (g Git) GetIssues(ctx context.Context, repo *entities.EnrichedRepo, ids []string) ([]entities.Issue, error) {
	intArray := make([]int, len(ids))
	for i, str := range ids {
		num, err := strconv.Atoi(str)
		if err != nil {
			return nil, err
		}
		intArray[i] = num
	}
	issues, _, err := g.c.Issues.ListProjectIssues(repo.ID, &gitlab.ListProjectIssuesOptions{IIDs: &intArray})
	if err != nil {
		return nil, err
	}

	var issueDetails []entities.Issue
	for _, issue := range issues {
		issueDetails = append(issueDetails, entities.Issue{
			IssueKey:     strconv.Itoa(issue.IID),
			IssueSummary: issue.Title,
			IssueStatus:  issue.State,
			IssueType:    *issue.IssueType,
			Category:     g.extractIssueCategory(*issue),
			WebURL:       issue.WebURL,
		})
	}

	return issueDetails, nil
}

func (g Git) extractIssueCategory(gi gitlab.Issue) entities.IssueCategory {
	for _, gil := range g.issueLabelsForFeature {
		for _, label := range gi.Labels {
			if strings.EqualFold(label, gil) {
				return entities.CLOSED_FEATURE
			}
		}
	}

	for _, gil := range g.issueLabelsForBug {
		for _, label := range gi.Labels {
			if strings.EqualFold(label, gil) {
				return entities.FIXED_BUG
			}
		}
	}

	return entities.CLOSED_FEATURE
}

func (g Git) GetKnownIssues(ctx context.Context, repo *entities.EnrichedRepo) ([]entities.Issue, error) {
	fmt.Printf("retrieving known issues using GitLab has not been implemented yet\n")
	return nil, nil
}

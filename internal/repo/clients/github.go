package clients

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v62/github"
	"github.com/happyagosmith/jig/internal/entities"
	"golang.org/x/oauth2"
)

type GitHub struct {
	c *github.Client
}

// NewGitHub creates a GitHub client using a personal access token or GITHUB_TOKEN.
func NewGitHub(token string) (GitHub, error) {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	httpClient := oauth2.NewClient(context.Background(), ts)
	return GitHub{c: github.NewClient(httpClient)}, nil
}

// NewGitHubApp creates a GitHub client authenticated as a GitHub App installation.
// appID is the App's numeric ID, installationID is the installation's numeric ID,
// and privateKeyPEM is the PEM-encoded private key of the App.
// If baseURL is non-empty, the client points to a GitHub Enterprise instance.
func NewGitHubApp(appID, installationID int64, privateKeyPEM []byte, baseURL string) (GitHub, error) {
	itr, err := ghinstallation.New(http.DefaultTransport, appID, installationID, privateKeyPEM)
	if err != nil {
		return GitHub{}, fmt.Errorf("failed to create GitHub App transport: %w", err)
	}
	if baseURL != "" {
		apiURL := strings.TrimRight(baseURL, "/") + "/api/v3/"
		itr.BaseURL = apiURL
		c, err := github.NewClient(&http.Client{Transport: itr}).WithEnterpriseURLs(apiURL, apiURL)
		if err != nil {
			return GitHub{}, err
		}
		return GitHub{c: c}, nil
	}
	return GitHub{c: github.NewClient(&http.Client{Transport: itr})}, nil
}

// NewGitHubWithBaseURL creates a GitHub client pointing at a custom base URL (useful for tests).
func NewGitHubWithBaseURL(token, baseURL string) (GitHub, error) {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	httpClient := oauth2.NewClient(context.Background(), ts)
	c, err := github.NewClient(httpClient).WithAuthToken(token).WithEnterpriseURLs(baseURL, baseURL)
	if err != nil {
		return GitHub{}, err
	}
	return GitHub{c: c}, nil
}

func splitOwnerRepo(gitRepoID string) (string, string, error) {
	parts := strings.SplitN(gitRepoID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("gitRepoID must be in the format owner/repo, got %q", gitRepoID)
	}
	return parts[0], parts[1], nil
}

func (g GitHub) GetCommits(id, from, to string) ([]entities.RepoRecord, error) {
	owner, repo, err := splitOwnerRepo(id)
	if err != nil {
		return nil, err
	}

	comparison, _, err := g.c.Repositories.CompareCommits(context.Background(), owner, repo, from, to, nil)
	if err != nil {
		return nil, err
	}

	var commits []entities.RepoRecord
	for _, c := range comparison.Commits {
		shortID := c.GetSHA()
		if len(shortID) > 8 {
			shortID = shortID[:8]
		}
		createdAt := c.GetCommit().GetCommitter().GetDate().Time
		commits = append(commits, entities.RepoRecord{
			ID:        c.GetSHA(),
			ShortID:   shortID,
			Title:     firstLine(c.GetCommit().GetMessage()),
			Message:   c.GetCommit().GetMessage(),
			CreatedAt: &createdAt,
			WebURL:    c.GetHTMLURL(),
			Origin:    "commit",
		})
	}

	return commits, nil
}

func (g GitHub) GetMergeRequests(id, targetBranch string, commits []entities.RepoRecord) ([]entities.RepoRecord, error) {
	if len(commits) == 0 {
		return nil, nil
	}

	owner, repo, err := splitOwnerRepo(id)
	if err != nil {
		return nil, err
	}

	lookForCommit := map[string]bool{}
	for _, c := range commits {
		lookForCommit[c.ID] = true
	}

	opts := &github.PullRequestListOptions{
		State: "closed",
		Base:  targetBranch,
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var result []entities.RepoRecord
	for {
		prs, resp, err := g.c.PullRequests.List(context.Background(), owner, repo, opts)
		if err != nil {
			return nil, err
		}

		for _, pr := range prs {
			if pr.GetMergeCommitSHA() == "" {
				continue
			}
			if !lookForCommit[pr.GetMergeCommitSHA()] {
				continue
			}
			mergedAt := pr.GetMergedAt().Time
			result = append(result, entities.RepoRecord{
				ID:        strconv.Itoa(int(pr.GetID())),
				ShortID:   strconv.Itoa(pr.GetNumber()),
				Title:     pr.GetTitle(),
				Message:   pr.GetBody(),
				CreatedAt: &mergedAt,
				WebURL:    pr.GetHTMLURL(),
				Origin:    "merge_request",
			})
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return result, nil
}

func (g GitHub) GetRepoURL(id string) (string, error) {
	owner, repo, err := splitOwnerRepo(id)
	if err != nil {
		return "", err
	}

	r, _, err := g.c.Repositories.Get(context.Background(), owner, repo)
	if err != nil {
		return "", err
	}

	return r.GetHTMLURL(), nil
}

func (g GitHub) GetReleaseURL(id, tag string) (string, error) {
	owner, repo, err := splitOwnerRepo(id)
	if err != nil {
		return "", err
	}

	r, _, err := g.c.Repositories.GetReleaseByTag(context.Background(), owner, repo, tag)
	if err != nil {
		return "", err
	}

	return r.GetHTMLURL(), nil
}

func (g GitHub) GetIssues(ctx context.Context, repo *entities.EnrichedRepo, ids []string) ([]entities.Issue, error) {
	owner, repoName, err := splitOwnerRepo(repo.ID)
	if err != nil {
		return nil, err
	}

	var issues []entities.Issue
	for _, id := range ids {
		num, err := strconv.Atoi(id)
		if err != nil {
			return nil, fmt.Errorf("invalid issue number %q: %w", id, err)
		}

		issue, _, err := g.c.Issues.Get(ctx, owner, repoName, num)
		if err != nil {
			return nil, err
		}

		issues = append(issues, entities.Issue{
			IssueKey:     strconv.Itoa(issue.GetNumber()),
			IssueSummary: issue.GetTitle(),
			IssueStatus:  issue.GetState(),
			IssueType:    "issue",
			Category:     extractGitHubIssueCategory(issue),
			WebURL:       issue.GetHTMLURL(),
		})
	}

	return issues, nil
}

func (g GitHub) GetKnownIssues(ctx context.Context, repo *entities.EnrichedRepo) ([]entities.Issue, error) {
	fmt.Printf("retrieving known issues using GitHub has not been implemented yet\n")
	return nil, nil
}

func extractGitHubIssueCategory(issue *github.Issue) entities.IssueCategory {
	for _, label := range issue.Labels {
		switch strings.ToLower(label.GetName()) {
		case "bug":
			return entities.FIXED_BUG
		case "feature", "enhancement":
			return entities.CLOSED_FEATURE
		}
	}
	return entities.CLOSED_FEATURE
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

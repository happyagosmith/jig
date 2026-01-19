package entities

import "context"

type IssuesTracker interface {
	GetIssues(ctx context.Context, repo *Repo, ids []string) ([]Issue, error)
	GetKnownIssues(ctx context.Context, repo *Repo) ([]Issue, error)
}

package entities

import "context"

type IssuesTracker interface {
	GetIssues(ctx context.Context, repo *EnrichedRepo, ids []string) ([]Issue, error)
	GetKnownIssues(ctx context.Context, repo *EnrichedRepo) ([]Issue, error)
}

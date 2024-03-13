package entities

type IssuesTracker interface {
	GetIssues(repo *Repo, ids []string) ([]Issue, error)
	GetKnownIssues(repo *Repo) ([]Issue, error)
}

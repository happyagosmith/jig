package entities

type RepoClient interface {
	GetCommits(id, from, to string) ([]RepoRecord, error)
	GetMergeRequests(id, targetBranch string, commits []RepoRecord) ([]RepoRecord, error)
	GetReleaseURL(id, tag string) (string, error)
	GetRepoURL(id string) (string, error)
}

type RepoTracker interface {
	RepoClient
	IssuesTracker
}

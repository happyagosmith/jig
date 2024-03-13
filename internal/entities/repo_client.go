package entities

type Repoclient interface {
	GetCommits(id, from, to string) ([]Commit, error)
	GetReleaseURL(id, tag string) (string, error)
	GetRepoURL(id string) (string, error)
}

type Repotracker interface {
	Repoclient
	IssuesTracker
}

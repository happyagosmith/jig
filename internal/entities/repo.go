package entities

type Repo struct {
	Label         string         `yaml:"label,omitempty"`
	ServiceName   string         `yaml:"serviceName,omitempty"`
	ID            string         `yaml:"gitRepoID,omitempty"`
	FromTag       string         `yaml:"previousVersion,omitempty"`
	ToTag         string         `yaml:"version,omitempty"`
	CheckTag      string         `yaml:"checkVersion,omitempty"`
	GitRepoURL    string         `yaml:"gitRepoURL,omitempty"`
	GitReleaseURL string         `yaml:"gitReleaseURL,omitempty"`
	Project       string         `yaml:"jiraProject,omitempty"`
	Component     string         `yaml:"jiraComponent,omitempty"`
	ParsedCommits []ParsedCommit `yaml:"extractedKeys,omitempty"`
	HasBreaking   bool           `yaml:"hasBreaking,omitempty"`
	HasNewFeature bool           `yaml:"hasNewFeature,omitempty"`
	HasBugFixed   bool           `yaml:"hasBugFixed,omitempty"`
}

type Repoparser interface {
	Parse([]Commit) ([]ParsedCommit, error)
}

package entities

import "fmt"

type Repo struct {
	Label            string         `yaml:"label,omitempty"`
	ServiceName      string         `yaml:"serviceName,omitempty"`
	ID               string         `yaml:"gitRepoID,omitempty"`
	FromTag          string         `yaml:"previousVersion,omitempty"`
	ToTag            string         `yaml:"version,omitempty"`
	CheckTag         string         `yaml:"checkVersion,omitempty"`
	Project          string         `yaml:"jiraProject,omitempty"`
	Component        string         `yaml:"jiraComponent,omitempty"`
	GitRepoURL       string         `yaml:"gitRepoURL,omitempty"`
	GitReleaseURL    string         `yaml:"gitReleaseURL,omitempty"`
	CustomAttributes map[string]any `yaml:"customAttributes,omitempty"`
}

type EnrichedRepo struct {
	Repo          `yaml:",inline"`
	ParsedCommits []ParsedRepoRecord `yaml:"extractedKeys,omitempty"`
	HasBreaking   bool               `yaml:"hasBreaking,omitempty"`
	HasNewFeature bool               `yaml:"hasNewFeature,omitempty"`
	HasBugFixed   bool               `yaml:"hasBugFixed,omitempty"`
}

func (r Repo) String() string {
	return fmt.Sprintf("repo %s from %s to %s (id: %s)\n", r.Label, r.FromTag, r.ToTag, r.ID)
}

type RepoService interface {
	GetParsedRecords(id, from, to, mrTargetBranch string) ([]ParsedRepoRecord, error)
	GetReleaseURL(id, tag string) (string, error)
	GetRepoURL(id string) (string, error)
}

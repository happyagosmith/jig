package model

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/happyagosmith/jig/internal/git"
	"github.com/happyagosmith/jig/internal/parsers"
	"gopkg.in/yaml.v2"
)

type Repo struct {
	Label         string             `yaml:"label,omitempty"`
	ServiceName   string             `yaml:"serviceName,omitempty"`
	ID            string             `yaml:"gitRepoID,omitempty"`
	FromCommit    string             `yaml:"fromCommit,omitempty"`
	ToCommit      string             `yaml:"toCommit,omitempty"`
	FromTag       string             `yaml:"previousVersion,omitempty"`
	ToTag         string             `yaml:"version,omitempty"`
	Project       string             `yaml:"jiraProject,omitempty"`
	Component     string             `yaml:"jiraComponent,omitempty"`
	CommitDetails []git.CommitDetail `yaml:"extractedKeys,omitempty"`
	HasBreaking   bool               `yaml:"hasBreaking"`
	HasNewFeature bool               `yaml:"hasNewFeature"`
	HasBugFixed   bool               `yaml:"hasBugFixed"`
}

type Conf struct {
	GitRepos []Repo `yaml:"services"`
}

type ExtractedIssue struct {
	IssueKey         string                   `yaml:"issueKey,omitempty"`
	IssueSummary     string                   `yaml:"issueSummary,omitempty"`
	IssueType        string                   `yaml:"issueType,omitempty"`
	IssueStatus      string                   `yaml:"issueStatus,omitempty"`
	IssueTrackerType parsers.IssueTrackerType `yaml:"issueTrackerType"`
	Category         CategoryType             `yaml:"category"`
	IsBreakingChange bool                     `yaml:"isBreakingChange"`
}

func (i ExtractedIssue) String() string {
	return fmt.Sprintf("key %s, issue type %s", i.IssueKey, i.Category)
}

type GeneratedValues struct {
	Features       map[string][]ExtractedIssue `yaml:"features"`
	Bugs           map[string][]ExtractedIssue `yaml:"bugs"`
	KnownIssues    map[string][]ExtractedIssue `yaml:"knownIssues"`
	BreakingChange map[string][]ExtractedIssue `yaml:"breakingChange"`
	GitRepos       []Repo                      `yaml:"gitRepos"`
}

type Model struct {
	gValues       GeneratedValues
	conf          Conf
	issueTrackers []IssuesTracker
	modelMap      map[string]any
}

type ModelOpt func(*Model)

func WithIssueTracker(it IssuesTracker) ModelOpt {
	return func(m *Model) {
		if it != nil {
			m.issueTrackers = append(m.issueTrackers, it)
		}
	}
}

func New(values []byte, opts ...ModelOpt) (*Model, error) {
	var model map[string]any
	err := yaml.Unmarshal(values, &model)
	if err != nil {
		return nil, err
	}

	var conf Conf
	err = yaml.Unmarshal(values, &conf)
	if err != nil {
		panic(err.Error())
	}

	m := Model{
		conf:     conf,
		modelMap: model,
		gValues: GeneratedValues{
			Features:       map[string][]ExtractedIssue{},
			Bugs:           map[string][]ExtractedIssue{},
			KnownIssues:    map[string][]ExtractedIssue{},
			BreakingChange: map[string][]ExtractedIssue{}},
	}

	for _, o := range opts {
		o(&m)
	}

	return &m, nil
}

func (m *Model) EnrichWithGit(URL, token, issuePattern, customPattern string) error {
	if URL == "" || token == "" {
		return fmt.Errorf("git URL and token are required")
	}
	if len(m.conf.GitRepos) == 0 {
		fmt.Printf("no git repos to process\n")
		return nil
	}
	git, err := git.NewClient(URL, token, issuePattern, customPattern)
	if err != nil {
		return err
	}

	for _, repo := range m.conf.GitRepos {
		if repo.FromTag != "" && repo.FromTag == repo.ToTag {
			fmt.Printf("same tag %s set in repo.FromTag and repo.ToTag for repo %s. Nothing changed \n", repo.FromTag, repo.Label)
			continue
		}

		if repo.FromCommit != "" && repo.FromCommit == repo.ToCommit {
			fmt.Printf("same commit %s set in repo.FromCommit and repo.ToCommit for repo %s. Nothing changed \n", repo.FromCommit, repo.Label)

			continue
		}

		fc := repo.FromTag
		if fc == "" {
			fc = repo.FromCommit
		}

		tc := repo.ToTag
		if tc == "" {
			tc = repo.ToCommit
		}

		fmt.Printf("\nprocessing repo \"%s\" from \"%s\" to \"%s\"\n", repo.Label, fc, tc)

		cds, err := git.ExtractCommits(repo.ID, fc, tc)
		if err != nil {
			return err
		}

		m.addCommitDetails(repo, cds)
	}

	return nil
}

type IssuesTracker interface {
	GetIssues(cds []git.CommitDetail) ([]ExtractedIssue, error)
	GetKnownIssues(project, component string) ([]ExtractedIssue, error)
	Type() parsers.IssueTrackerType
}

type CategoryType int

const (
	CLOSED_FEATURE CategoryType = iota
	FIXED_BUG
	SUB_TASK
	OTHER
)

func (i CategoryType) String() string {
	return []string{"CLOSED_FEATURE", "FIXED_BUG", "SUB_TASK", "OTHER"}[i]
}

func (i CategoryType) MarshalYAML() (interface{}, error) {
	return i.String(), nil
}

func (m *Model) EnrichWithIssueTrackers() error {
	for i := range m.gValues.GitRepos {
		repo := &m.gValues.GitRepos[i]
		err := m.enrichRepoWithIssueTracker(repo)
		if err != nil {
			return err
		}

		sv, _ := computeSemanticVersion(repo.FromTag, repo.HasBreaking, repo.HasNewFeature, repo.HasBugFixed)
		fmt.Printf("\ncurrent version for the repo \"%s\" is: %s, suggested version \"%s\" instead\n", repo.Label, repo.FromTag, sv)
	}

	return nil
}

func (m *Model) enrichRepoWithIssueTracker(repo *Repo) error {
	for _, issuesTracker := range m.issueTrackers {
		fmt.Printf("\nretrieving issues info from the issues tracker \"%s\" for the repo \"%s\"\n", issuesTracker.Type(), repo.Label)
		issues, err := issuesTracker.GetIssues(repo.CommitDetails)
		if err != nil {
			return err
		}
		hasBreaking, hasNewFeature, hasBugFixed := m.addFoundIssues(repo.Label, issues, issuesTracker.Type())
		repo.HasBreaking = repo.HasBreaking || hasBreaking
		repo.HasNewFeature = repo.HasNewFeature || hasNewFeature
		repo.HasBugFixed = repo.HasBugFixed || hasBugFixed

		if repo.Project == "" {
			fmt.Printf("\nknown issues not retrieved. project not set for the repo \"%s\"\n", repo.Label)
			continue
		}

		knownIssues, err := issuesTracker.GetKnownIssues(repo.Project, repo.Component)
		if err != nil {
			return err
		}

		m.addKnownIssues(repo.Label, knownIssues, issuesTracker.Type())
	}

	return nil
}

func (m *Model) addFoundIssues(label string, issues []ExtractedIssue, it parsers.IssueTrackerType) (bool, bool, bool) {
	var hasBreaking, hasNewFeature, hasBugFixed bool

	for _, issue := range issues {
		fmt.Printf("analysing %s\n", issue.String())
		if issue.Category == SUB_TASK {
			fmt.Print("subTask not added\n")
			continue
		}
		if issue.IsBreakingChange {
			m.gValues.BreakingChange[label] = append(m.gValues.BreakingChange[label], issue)
			fmt.Print("added as Breaking Change\n")
			hasBreaking = true
		}
		if issue.Category == CLOSED_FEATURE {
			m.gValues.Features[label] = append(m.gValues.Features[label], issue)
			fmt.Print("added as feature\n")
			hasNewFeature = true
			continue
		}
		if issue.Category == FIXED_BUG {
			m.gValues.Bugs[label] = append(m.gValues.Bugs[label], issue)
			fmt.Print("added as bug\n")
			hasBugFixed = true
			continue
		}
	}

	return hasBreaking, hasNewFeature, hasBugFixed
}

func (m *Model) addKnownIssues(label string, issues []ExtractedIssue, it parsers.IssueTrackerType) {
	for _, issue := range issues {
		issue.IssueTrackerType = it
		m.gValues.KnownIssues[label] = append(m.gValues.KnownIssues[label], issue)
		fmt.Printf("added %s\n", issue.String())
	}
}

func (m *Model) addCommitDetails(repo Repo, cds []git.CommitDetail) {
	repo.CommitDetails = cds
	m.gValues.GitRepos = append(m.gValues.GitRepos, repo)

	for _, issue := range cds {
		if issue.IssueTracker != parsers.NONE {
			continue
		}
		fmt.Printf("analysing %s\n", issue.String())
		ei := ExtractedIssue{
			IssueKey:         issue.IssueKey,
			IssueSummary:     issue.Summary,
			IssueType:        issue.Category.String(),
			IssueTrackerType: parsers.NONE,
			IsBreakingChange: issue.IsBreaking,
		}
		if issue.IsBreaking {
			m.gValues.BreakingChange[repo.Label] = append(m.gValues.BreakingChange[repo.Label], ei)
			fmt.Print("added as Breaking Change\n")
			repo.HasBreaking = true
		}
		if issue.Category == parsers.FEATURE {
			m.gValues.Features[repo.Label] = append(m.gValues.Features[repo.Label], ei)
			fmt.Print("added as feature\n")
			repo.HasNewFeature = true
			continue
		}
		if issue.Category == parsers.BUG_FIX {
			m.gValues.Bugs[repo.Label] = append(m.gValues.Bugs[repo.Label], ei)
			fmt.Print("added as bug\n")
			repo.HasBugFixed = true
			continue
		}
	}
}

func computeSemanticVersion(currentVersion string, hasBreaking, hasNewFeature, hasBugFixed bool) (string, error) {
	v := strings.Split(currentVersion, ".")
	if len(v) < 3 {
		return "", fmt.Errorf("the current version do not have a semantic version format")
	}

	ma, err := strconv.Atoi(v[0])
	if err != nil {
		return currentVersion, fmt.Errorf("the current version do not have a semantic version format")
	}

	mi, err := strconv.Atoi(v[1])
	if err != nil {
		return currentVersion, fmt.Errorf("the current version do not have a semantic version format")
	}

	p, err := strconv.Atoi(v[2])
	if err != nil {
		return currentVersion, fmt.Errorf("the current version do not have a semantic version format")
	}

	if hasBreaking {
		ma++
		return fmt.Sprintf("%d.%d.%d", ma, mi, p), nil
	}

	if hasNewFeature {
		mi++
		return fmt.Sprintf("%d.%d.%d", ma, mi, p), nil
	}

	if hasBugFixed {
		p++
		return fmt.Sprintf("%d.%d.%d", ma, mi, p), nil
	}

	return fmt.Sprintf("%d.%d.%d", ma, mi, p), nil
}

func (m *Model) Map() map[string]any {
	m.modelMap["generatedValues"] = m.gValues
	return m.modelMap
}

func (m *Model) Yaml() ([]byte, error) {
	return yaml.Marshal(m.Map())
}

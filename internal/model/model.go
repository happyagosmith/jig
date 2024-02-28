package model

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/happyagosmith/jig/internal/git"
	"github.com/happyagosmith/jig/internal/parsers"
	"github.com/happyagosmith/jig/internal/utils"
	"gopkg.in/yaml.v2"
)

type Repo struct {
	Label         string             `yaml:"label,omitempty"`
	ServiceName   string             `yaml:"serviceName,omitempty"`
	ID            string             `yaml:"gitRepoID,omitempty"`
	FromTag       string             `yaml:"previousVersion,omitempty"`
	ToTag         string             `yaml:"version,omitempty"`
	CheckTag      string             `yaml:"checkVersion,omitempty"`
	Project       string             `yaml:"jiraProject,omitempty"`
	Component     string             `yaml:"jiraComponent,omitempty"`
	CommitDetails []git.CommitDetail `yaml:"extractedKeys,omitempty"`
	HasBreaking   bool               `yaml:"hasBreaking,omitempty"`
	HasNewFeature bool               `yaml:"hasNewFeature,omitempty"`
	HasBugFixed   bool               `yaml:"hasBugFixed,omitempty"`
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
	GValues       *GeneratedValues `yaml:"generatedValues,omitempty"`
	GitRepos      []Repo           `yaml:"services"`
	issueTrackers []IssuesTracker
	vcs           VCS
	y             *utils.Yaml
}

type ModelOpt func(*Model)

type VCS interface {
	ExtractCommits(id, from, to string) ([]git.CommitDetail, error)
}

func WithVCS(vcs VCS) ModelOpt {
	return func(m *Model) {
		if vcs != nil {
			m.vcs = vcs
		}
	}
}

func WithIssueTracker(it IssuesTracker) ModelOpt {
	return func(m *Model) {
		if it != nil {
			m.issueTrackers = append(m.issueTrackers, it)
		}
	}
}

func New(values []byte, opts ...ModelOpt) (*Model, error) {
	var m Model
	err := yaml.Unmarshal(values, &m)
	if err != nil {
		panic(err.Error())
	}
	m.GValues = nil

	yaml, err := utils.NewYaml(values)
	if err != nil {
		panic(err.Error())
	}
	err = yaml.Delete("generatedValues")
	if err != nil {
		panic(err.Error())
	}

	m.y = yaml

	for _, o := range opts {
		o(&m)
	}

	return &m, nil
}

func (m *Model) SetVersions(rootPath string) error {
	for i, repo := range m.GitRepos {
		if !strings.HasPrefix(repo.CheckTag, "@") {
			continue
		}

		p := strings.Split(repo.CheckTag, ":")
		if len(p) < 2 {
			continue
		}

		path := strings.Join([]string{rootPath, strings.TrimPrefix(p[0], "@")}, "/")
		dataYaml, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		y, err := utils.NewYaml(dataYaml)
		if err != nil {
			return err
		}

		wantTag, err := y.GetValue(p[1])
		if err != nil {
			return err
		}

		if wantTag == repo.ToTag {
			continue
		}

		repo.FromTag = repo.ToTag
		repo.ToTag = wantTag

		m.GitRepos[i] = repo
	}

	return nil
}

func (m *Model) EnrichWithGit() error {
	if len(m.GitRepos) == 0 {
		fmt.Printf("no git repos to process\n")
		return nil
	}

	if m.GValues == nil {
		m.GValues = &GeneratedValues{}
	}

	m.GValues.GitRepos = []Repo{}
	for _, repo := range m.GitRepos {
		if repo.FromTag != "" && repo.FromTag == repo.ToTag {
			fmt.Printf("same tag %s set in repo.FromTag and repo.ToTag for repo %s. Nothing changed \n", repo.FromTag, repo.Label)
			continue
		}

		fc := repo.FromTag
		tc := repo.ToTag

		fmt.Printf("\nprocessing repo \"%s\" from \"%s\" to \"%s\"\n", repo.Label, fc, tc)

		cds, err := m.vcs.ExtractCommits(repo.ID, fc, tc)
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

func (ct CategoryType) String() string {
	return []string{"CLOSED_FEATURE", "FIXED_BUG", "SUB_TASK", "OTHER"}[ct]
}

func (ct CategoryType) MarshalYAML() (interface{}, error) {
	return ct.String(), nil
}

func (ct *CategoryType) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}

	switch strings.ToLower(s) {
	case "closed_feature":
		*ct = CLOSED_FEATURE
	case "fixed_bug":
		*ct = FIXED_BUG
	case "sub_task":
		*ct = SUB_TASK
	case "other":
		*ct = OTHER
	default:
		return fmt.Errorf("invalid CategoryType %q", s)
	}

	return nil
}

func (m *Model) EnrichWithIssueTrackers() error {
	if m.GValues == nil {
		m.GValues = &GeneratedValues{}
	}
	m.GValues.Features = map[string][]ExtractedIssue{}
	m.GValues.Bugs = map[string][]ExtractedIssue{}
	m.GValues.KnownIssues = map[string][]ExtractedIssue{}
	m.GValues.BreakingChange = map[string][]ExtractedIssue{}

	for i := range m.GValues.GitRepos {
		repo := &m.GValues.GitRepos[i]
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
			m.GValues.BreakingChange[label] = append(m.GValues.BreakingChange[label], issue)
			fmt.Print("added as Breaking Change\n")
			hasBreaking = true
		}
		if issue.Category == CLOSED_FEATURE {
			m.GValues.Features[label] = append(m.GValues.Features[label], issue)
			fmt.Print("added as feature\n")
			hasNewFeature = true
			continue
		}
		if issue.Category == FIXED_BUG {
			m.GValues.Bugs[label] = append(m.GValues.Bugs[label], issue)
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
		m.GValues.KnownIssues[label] = append(m.GValues.KnownIssues[label], issue)
		fmt.Printf("added %s\n", issue.String())
	}
}

func (m *Model) addCommitDetails(repo Repo, cds []git.CommitDetail) {
	repo.CommitDetails = cds
	m.GValues.GitRepos = append(m.GValues.GitRepos, repo)

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
			m.GValues.BreakingChange[repo.Label] = append(m.GValues.BreakingChange[repo.Label], ei)
			fmt.Print("added as Breaking Change\n")
			repo.HasBreaking = true
		}
		if issue.Category == parsers.FEATURE {
			m.GValues.Features[repo.Label] = append(m.GValues.Features[repo.Label], ei)
			fmt.Print("added as feature\n")
			repo.HasNewFeature = true
			continue
		}
		if issue.Category == parsers.BUG_FIX {
			m.GValues.Bugs[repo.Label] = append(m.GValues.Bugs[repo.Label], ei)
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

func (m *Model) Yaml() ([]byte, error) {
	yb, err := yaml.Marshal(m)
	if err != nil {
		return nil, err
	}

	y, err := utils.NewYaml(yb)
	if err != nil {
		return nil, err
	}

	if err := m.y.Merge(y); err != nil {
		return nil, err
	}

	return m.y.Bytes()
}

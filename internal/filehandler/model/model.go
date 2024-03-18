package model

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/happyagosmith/jig/internal/entities"
	yamlfile "github.com/happyagosmith/jig/internal/filehandler/yaml"
	"gopkg.in/yaml.v2"
)

type GeneratedValues struct {
	Features       map[string][]entities.ExtractedIssue `yaml:"features"`
	Bugs           map[string][]entities.ExtractedIssue `yaml:"bugs"`
	KnownIssues    map[string][]entities.ExtractedIssue `yaml:"knownIssues"`
	BreakingChange map[string][]entities.ExtractedIssue `yaml:"breakingChange"`
	GitRepos       []entities.Repo                      `yaml:"gitRepos"`
}

type Model struct {
	GValues       *GeneratedValues `yaml:"generatedValues,omitempty"`
	GitRepos      []entities.Repo  `yaml:"services"`
	issueTrackers []struct {
		label string
		it    entities.IssuesTracker
	}
	repoparser entities.Repoparser
	repoclient entities.Repoclient
	y          *yamlfile.Yaml
}

type ModelOpt func(*Model)

func WithRepoParser(repoparser entities.Repoparser) ModelOpt {
	return func(m *Model) {
		if repoparser != nil {
			m.repoparser = repoparser
		}
	}
}

func WithRepoClient(repoclient entities.Repoclient) ModelOpt {
	return func(m *Model) {
		if repoclient != nil {
			m.repoclient = repoclient
		}
	}
}

func WithIssueTracker(label string, it entities.IssuesTracker) ModelOpt {
	return func(m *Model) {
		if label != "" {
			m.issueTrackers = append(m.issueTrackers, struct {
				label string
				it    entities.IssuesTracker
			}{it: it, label: label})
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

	yaml, err := yamlfile.NewYaml(values)
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

func (m *Model) UpdateWithReposVersions(rootPath string) error {
	for i, repo := range m.GitRepos {
		if !strings.HasPrefix(repo.CheckTag, "@") {
			continue
		}

		p := strings.Split(repo.CheckTag, ":")
		if len(p) < 2 {
			continue
		}

		path := strings.TrimPrefix(p[0], "@")
		if strings.HasPrefix(path, ".") {
			path = filepath.Join(rootPath, path)
		}

		dataYaml, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		y, err := yamlfile.NewYaml(dataYaml)
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

func (m *Model) UpdateWithReposInfos() error {
	if m.repoclient == nil {
		return fmt.Errorf("vcs not set")
	}
	for i, repo := range m.GitRepos {

		rUrl, err := m.repoclient.GetRepoURL(repo.ID)
		if err != nil {
			return err
		}

		vUrl, err := m.repoclient.GetReleaseURL(repo.ID, repo.ToTag)
		if err != nil {
			return err
		}

		m.GitRepos[i].GitRepoURL = rUrl
		m.GitRepos[i].GitReleaseURL = vUrl
	}

	return nil
}

func (m *Model) EnrichWithRepos() error {
	if len(m.GitRepos) == 0 {
		fmt.Printf("no git repos to process\n")
		return nil
	}

	if m.GValues == nil {
		m.GValues = &GeneratedValues{}
		m.GValues.Features = map[string][]entities.ExtractedIssue{}
		m.GValues.Bugs = map[string][]entities.ExtractedIssue{}
		m.GValues.KnownIssues = map[string][]entities.ExtractedIssue{}
		m.GValues.BreakingChange = map[string][]entities.ExtractedIssue{}
	}

	m.GValues.GitRepos = []entities.Repo{}
	for _, repo := range m.GitRepos {
		if repo.FromTag != "" && repo.FromTag == repo.ToTag {
			fmt.Printf("same tag %s set in repo.FromTag and repo.ToTag for repo %s. Nothing changed \n", repo.FromTag, repo.Label)
			continue
		}

		fc := repo.FromTag
		tc := repo.ToTag

		fmt.Printf("\nprocessing repo \"%s\" from \"%s\" to \"%s\"\n", repo.Label, fc, tc)

		commits, err := m.repoclient.GetCommits(repo.ID, fc, tc)
		if err != nil {
			return err
		}

		pcommits, err := m.repoparser.Parse(commits)
		if err != nil {
			return err
		}

		m.addParsedCommits(repo, pcommits)
	}

	return nil
}

func (m *Model) EnrichWithIssueTrackers() error {
	if m.GValues == nil {
		m.GValues = &GeneratedValues{}
	}
	m.GValues.Features = map[string][]entities.ExtractedIssue{}
	m.GValues.Bugs = map[string][]entities.ExtractedIssue{}
	m.GValues.KnownIssues = map[string][]entities.ExtractedIssue{}
	m.GValues.BreakingChange = map[string][]entities.ExtractedIssue{}

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

func (m *Model) enrichRepoWithIssueTracker(repo *entities.Repo) error {
	for _, issuesTracker := range m.issueTrackers {
		keys := []string{}
		commits := map[string]entities.ParsedCommit{}

		for _, gc := range repo.ParsedCommits {
			if gc.ParsedIssueTracker != issuesTracker.label {
				continue
			}

			keys = append(keys, gc.ParsedKey)
			commits[gc.ParsedKey] = gc
		}

		if issuesTracker.it == nil && len(keys) != 0 {
			fmt.Printf("\nissues tracker implementantion not set for the type \"%s\"\n", issuesTracker.label)
			fmt.Printf("adding issues with only commit details \"%s\"\n", issuesTracker.label)
			m.addParsedCommitAsIssues(repo.Label, commits)
			continue
		}

		if len(keys) != 0 {
			fmt.Printf("\nretrieving issues info from the issues tracker \"%s\" for the repo \"%s\"\n", issuesTracker.label, repo.Label)
			issues, err := issuesTracker.it.GetIssues(repo, keys)
			if err != nil {
				return err
			}
			extractedIssues := make([]entities.ExtractedIssue, 0, len(issues))
			for _, issue := range issues {
				extractedIssues = append(extractedIssues, entities.ExtractedIssue{
					IssueTracker:  issuesTracker.label,
					IssueKey:      issue.IssueKey,
					IssueSummary:  issue.IssueSummary,
					IssueCategory: issue.Category,
					Issue:         issue,
					ParsedCommit:  commits[issue.IssueKey],
				})
			}
			hasBreaking, hasNewFeature, hasBugFixed := m.addFoundIssues(repo.Label, extractedIssues)
			repo.HasBreaking = repo.HasBreaking || hasBreaking
			repo.HasNewFeature = repo.HasNewFeature || hasNewFeature
			repo.HasBugFixed = repo.HasBugFixed || hasBugFixed
		}

		if repo.Project == "" {
			fmt.Printf("\nknown issues not retrieved. project not set for the repo \"%s\"\n", repo.Label)
			continue
		}

		knownIssues, err := issuesTracker.it.GetKnownIssues(repo)
		if err != nil {
			return err
		}

		m.addKnownIssues(repo.Label, knownIssues, issuesTracker.label)
	}

	return nil
}

func (m *Model) addFoundIssues(label string, issues []entities.ExtractedIssue) (bool, bool, bool) {
	var hasBreaking, hasNewFeature, hasBugFixed bool

	for _, issue := range issues {
		fmt.Printf("analysing %s\n", issue.String())
		if issue.Issue.Category == entities.SUB_TASK {
			fmt.Print("subTask not added\n")
			continue
		}
		if issue.ParsedCommit.IsBreakingChange {
			m.GValues.BreakingChange[label] = append(m.GValues.BreakingChange[label], issue)
			fmt.Print("added as Breaking Change\n")
			hasBreaking = true
		}
		if issue.Issue.Category == entities.CLOSED_FEATURE {
			m.GValues.Features[label] = append(m.GValues.Features[label], issue)
			fmt.Print("added as feature\n")
			hasNewFeature = true
			continue
		}
		if issue.Issue.Category == entities.FIXED_BUG {
			m.GValues.Bugs[label] = append(m.GValues.Bugs[label], issue)
			fmt.Print("added as bug\n")
			hasBugFixed = true
			continue
		}
	}

	return hasBreaking, hasNewFeature, hasBugFixed
}

func (m *Model) addParsedCommitAsIssues(label string, commits map[string]entities.ParsedCommit) (bool, bool, bool) {
	var hasBreaking, hasNewFeature, hasBugFixed bool

	for _, c := range commits {
		fmt.Printf("analysing %s\n", c.String())
		ei := entities.ExtractedIssue{
			IssueTracker: c.ParsedIssueTracker,
			IssueKey:     c.ParsedKey,
			IssueSummary: c.Summary,
			ParsedCommit: c}
		if c.ParsedCategory == entities.FEATURE {
			ei.IssueCategory = entities.CLOSED_FEATURE
			m.GValues.Features[label] = append(m.GValues.Features[label], ei)
			fmt.Print("added as feature\n")
			hasNewFeature = true
		}
		if c.ParsedCategory == entities.BUG_FIX {
			ei.IssueCategory = entities.FIXED_BUG
			m.GValues.Bugs[label] = append(m.GValues.Bugs[label], ei)
			fmt.Print("added as bug\n")
			hasBugFixed = true
		}
		if c.IsBreakingChange {
			m.GValues.BreakingChange[label] = append(m.GValues.BreakingChange[label], ei)
			fmt.Print("added as Breaking Change\n")
			hasBreaking = true
		}

	}

	return hasBreaking, hasNewFeature, hasBugFixed
}

func (m *Model) addKnownIssues(label string, issues []entities.Issue, it string) {
	for _, issue := range issues {

		m.GValues.KnownIssues[label] = append(m.GValues.KnownIssues[label], entities.ExtractedIssue{
			IssueKey:      issue.IssueKey,
			IssueSummary:  issue.IssueSummary,
			IssueCategory: issue.Category,
			IssueTracker:  it,
			Issue:         issue,
		})
		fmt.Printf("added %s\n", issue.String())
	}
}

func (m *Model) addParsedCommits(repo entities.Repo, cds []entities.ParsedCommit) {
	repo.ParsedCommits = cds
	m.GValues.GitRepos = append(m.GValues.GitRepos, repo)

	for _, issue := range cds {
		fmt.Printf("analysing %s\n", issue.String())
		ei := entities.ExtractedIssue{}
		ei.ParsedCommit = issue

		if issue.IsBreakingChange {
			m.GValues.BreakingChange[repo.Label] = append(m.GValues.BreakingChange[repo.Label], ei)
			fmt.Print("added as Breaking Change\n")
			repo.HasBreaking = true
		}
		if issue.ParsedCategory == entities.FEATURE {
			m.GValues.Features[repo.Label] = append(m.GValues.Features[repo.Label], ei)
			fmt.Print("added as feature\n")
			repo.HasNewFeature = true
			continue
		}
		if issue.ParsedCategory == entities.BUG_FIX {
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

	y, err := yamlfile.NewYaml(yb)
	if err != nil {
		return nil, err
	}

	if err := m.y.Merge(y); err != nil {
		return nil, err
	}

	return m.y.Bytes()
}

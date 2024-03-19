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
	repoService entities.RepoService
	y           *yamlfile.Yaml
}

type ModelOpt func(*Model)

func WithRepoService(repoService entities.RepoService) ModelOpt {
	return func(m *Model) {
		if repoService != nil {
			m.repoService = repoService
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
	if m.repoService == nil {
		return fmt.Errorf("vcs not set")
	}
	for i, repo := range m.GitRepos {

		rUrl, err := m.repoService.GetRepoURL(repo.ID)
		if err != nil {
			return err
		}

		vUrl, err := m.repoService.GetReleaseURL(repo.ID, repo.ToTag)
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
	}
	m.GValues.Features = map[string][]entities.ExtractedIssue{}
	m.GValues.Bugs = map[string][]entities.ExtractedIssue{}
	m.GValues.KnownIssues = map[string][]entities.ExtractedIssue{}
	m.GValues.BreakingChange = map[string][]entities.ExtractedIssue{}

	m.GValues.GitRepos = []entities.Repo{}
	for _, repo := range m.GitRepos {
		if repo.FromTag != "" && repo.FromTag == repo.ToTag {
			fmt.Printf("same tag %s set in repo.FromTag and repo.ToTag for repo %s. Nothing changed \n", repo.FromTag, repo.Label)
			continue
		}

		fc := repo.FromTag
		tc := repo.ToTag

		fmt.Printf("\nprocessing %s", repo.String())

		pRecords, err := m.repoService.GetParsedRecords(repo.ID, fc, tc, "")
		if err != nil {
			return err
		}

		repo.ParsedCommits = pRecords
		m.GValues.GitRepos = append(m.GValues.GitRepos, repo)
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
		fmt.Printf("\ncurrent version for the repo \"%s\" is: %s, suggested version \"%s\"\n", repo.Label, repo.FromTag, sv)
	}

	return nil
}

func (m *Model) enrichRepoWithIssueTracker(repo *entities.Repo) error {
	for _, issuesTracker := range m.issueTrackers {
		fmt.Printf("\nenriching repo %s with issues info from the issues tracker \"%s\"\n", repo.Label, issuesTracker.label)
		keys := []string{}
		commits := map[string]entities.ParsedRepoRecord{}

		for _, gc := range repo.ParsedCommits {
			if gc.ParsedIssueTracker != issuesTracker.label {
				continue
			}

			keys = append(keys, gc.ParsedKey)
			commits[gc.ParsedKey] = gc
		}

		if issuesTracker.it == nil {
			fmt.Printf("issues tracker implementation not set for the type \"%s\"\n", issuesTracker.label)
			fmt.Printf("adding issues with only commit details \"%s\"\n", issuesTracker.label)
			m.addParsedCommitAsIssues(repo.Label, commits)
			continue
		}

		if issuesTracker.it == nil {
			fmt.Printf("issues tracker implementation not set for the type \"%s\"\n", issuesTracker.label)
			continue
		}

		if len(keys) != 0 {
			fmt.Printf("retrieving issues info from the issues tracker \"%s\" for the repo \"%s\"\n", issuesTracker.label, repo.Label)
			issues, err := issuesTracker.it.GetIssues(repo, keys)
			if err != nil {
				return err
			}
			extractedIssues := make([]entities.ExtractedIssue, 0, len(issues))
			for _, issue := range issues {
				extractedIssues = append(extractedIssues, entities.ExtractedIssue{
					IssueTracker:     issuesTracker.label,
					IssueKey:         issue.IssueKey,
					IssueSummary:     issue.IssueSummary,
					IssueCategory:    issue.Category,
					Issue:            issue,
					ParsedRepoRecord: commits[issue.IssueKey],
				})
			}
			hasBreaking, hasNewFeature, hasBugFixed := m.addFoundIssues(repo.Label, extractedIssues)
			repo.HasBreaking = repo.HasBreaking || hasBreaking
			repo.HasNewFeature = repo.HasNewFeature || hasNewFeature
			repo.HasBugFixed = repo.HasBugFixed || hasBugFixed
		}

		knownIssues, err := issuesTracker.it.GetKnownIssues(repo)
		if err != nil {
			return err
		}

		if len(knownIssues) == 0 {
			fmt.Printf("no known issues retrieved from the issues tracker \"%s\" for the repo \"%s\"\n", issuesTracker.label, repo.Label)
			continue
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
		if issue.ParsedRepoRecord.IsBreakingChange {
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

func (m *Model) addParsedCommitAsIssues(label string, commits map[string]entities.ParsedRepoRecord) (bool, bool, bool) {
	var hasBreaking, hasNewFeature, hasBugFixed bool
	for _, c := range commits {
		fmt.Printf("analysing %s\n", c.ShortString())
		summary := c.ParsedSummary
		if summary == "" {
			summary = c.RepoRecord.Title
		}
		ei := entities.ExtractedIssue{
			IssueTracker:     c.ParsedIssueTracker,
			IssueKey:         c.ParsedKey,
			IssueSummary:     summary,
			ParsedRepoRecord: c}
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

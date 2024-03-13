package model_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/happyagosmith/jig/internal/entities"
	"github.com/happyagosmith/jig/internal/filehandler/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepoParser struct {
	mock.Mock
}

func (m *MockRepoParser) Parse(commits []entities.Commit) ([]entities.ParsedCommit, error) {
	args := m.Called(commits)
	return args.Get(0).([]entities.ParsedCommit), args.Error(1)
}

type MockGitLabClient struct {
	mock.Mock
}

func (m *MockGitLabClient) GetCommits(repoID, fromTag, toTag string) ([]entities.Commit, error) {
	args := m.Called(repoID, fromTag, toTag)
	return args.Get(0).([]entities.Commit), args.Error(1)
}

func (m *MockGitLabClient) GetReleaseURL(id, version string) (string, error) {
	args := m.Called(id)
	return args.Get(0).(string), args.Error(1)
}

func (m *MockGitLabClient) GetRepoURL(id string) (string, error) {
	args := m.Called(id)
	return args.Get(0).(string), args.Error(1)
}

func TestSetVersions(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "example.*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte("version: 1.0.1")); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	values := []byte(fmt.Sprintf(""+
		"services:\n"+
		"  - previousVersion: 0.0.0\n"+
		"    version: 1.0.0\n"+
		"    checkVersion: '@%s:version'\n"+
		"generatedValues:\n"+
		"  features: {}\n"+
		"  bugs: {}\n"+
		"  knownIssues: {}\n"+
		"  breakingChange: {}\n"+
		"  gitRepos: []\n", tmpfile.Name()))

	wantContent := fmt.Sprintf(""+
		"services:\n"+
		"  - previousVersion: 1.0.0\n"+
		"    version: 1.0.1\n"+
		"    checkVersion: '@%s:version'\n", tmpfile.Name())

	m, err := model.New(values)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err = m.UpdateWithReposVersions(os.TempDir())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	bytes, err := m.Yaml()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	assert.Equal(t, wantContent, string(bytes))
}

func TestEnrichWithGit(t *testing.T) {
	tests := []struct {
		name            string
		repoID          string
		previousVersion string
		version         string
		wantContent     string
	}{
		{
			name:            "Test EnrichWithGit",
			repoID:          "repo1",
			previousVersion: "0.0.0",
			version:         "1.0.0",
			wantContent: "" +
				"services:\n" +
				"  - label: label1\n" +
				"    gitRepoID: repo1\n" +
				"    previousVersion: 0.0.0\n" +
				"    version: 1.0.0\n" +
				"generatedValues:\n" +
				"  features: {}\n" +
				"  bugs: {}\n" +
				"  knownIssues: {}\n" +
				"  breakingChange: {}\n" +
				"  gitRepos:\n" +
				"    - label: label1\n" +
				"      gitRepoID: repo1\n" +
				"      previousVersion: 0.0.0\n" +
				"      version: 1.0.0\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockedCommtis := []entities.Commit{}

			mockGitLabClient := new(MockGitLabClient)
			mockGitLabClient.On("GetCommits", "repo1", "0.0.0", "1.0.0").Return(mockedCommtis, nil)

			mockRepoParser := new(MockRepoParser)
			mockRepoParser.On("Parse", mockedCommtis).Return([]entities.ParsedCommit{}, nil)

			values := []byte(fmt.Sprintf(`
services:
  - label: label1
    gitRepoID: %s
    previousVersion: %s
    version: %s
`, tt.repoID, tt.previousVersion, tt.version))

			m, err := model.New(values, model.WithRepoClient(mockGitLabClient), model.WithRepoParser(mockRepoParser))
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			err = m.EnrichWithRepos()

			mockGitLabClient.AssertExpectations(t)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(mockGitLabClient.Calls))

			bytes, err := m.Yaml()
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			assert.Equal(t, tt.wantContent, string(bytes))
		})
	}
}

type MockIssueTracker struct {
	mock.Mock
}

func (m *MockIssueTracker) GetIssues(_ *entities.Repo, keys []string) ([]entities.Issue, error) {
	args := m.Called(keys)
	return args.Get(0).([]entities.Issue), args.Error(1)
}

func (m *MockIssueTracker) GetKnownIssues(repo *entities.Repo) ([]entities.Issue, error) {
	args := m.Called(repo)
	return args.Get(0).([]entities.Issue), args.Error(1)
}

func TestEnrichWithIssueTrackers(t *testing.T) {
	tests := []struct {
		name            string
		parsedCommits   []entities.ParsedCommit
		issues          []entities.Issue
		knownIssues     []entities.Issue
		expectedContent string
		expectedError   error
	}{
		{
			name:          "Test EnrichWithIssueTrackers no issues",
			parsedCommits: []entities.ParsedCommit{{ParsedKey: "AAA-000", ParsedIssueTracker: "JIRA"}},
			issues:        []entities.Issue{},
			knownIssues:   []entities.Issue{},
			expectedContent: "services:\n" +
				"  - label: label1\n" +
				"    gitRepoID: repoID\n" +
				"    previousVersion: 0.0.0\n" +
				"    version: 1.0.0\n" +
				"    jiraProject: project\n" +
				"    jiraComponent: component\n" +
				"generatedValues:\n" +
				"  features: {}\n" +
				"  bugs: {}\n" +
				"  knownIssues: {}\n" +
				"  breakingChange: {}\n" +
				"  gitRepos:\n" +
				"    - label: label1\n" +
				"      gitRepoID: repoID\n" +
				"      previousVersion: 0.0.0\n" +
				"      version: 1.0.0\n" +
				"      jiraProject: project\n" +
				"      jiraComponent: component\n" +
				"      extractedKeys:\n" +
				"        - parsedKey: AAA-000\n" +
				"          parsedIssueTracker: JIRA\n",
			expectedError: nil,
		},
		{
			name:          "Test EnrichWithIssueTrackers with features",
			parsedCommits: []entities.ParsedCommit{{ParsedKey: "AAA-000", ParsedIssueTracker: "JIRA", IsBreakingChange: false}},
			issues: []entities.Issue{{
				IssueKey:     "AAA-000",
				IssueSummary: "summary",
				IssueType:    "type",
				IssueStatus:  "status",
				Category:     entities.CLOSED_FEATURE,
			},
			},
			knownIssues: []entities.Issue{},
			expectedContent: "services:\n" +
				"  - label: label1\n" +
				"    gitRepoID: repoID\n" +
				"    previousVersion: 0.0.0\n" +
				"    version: 1.0.0\n" +
				"    jiraProject: project\n" +
				"    jiraComponent: component\n" +
				"generatedValues:\n" +
				"  features:\n" +
				"    label1:\n" +
				"      - issueTracker: JIRA\n" +
				"        category: CLOSED_FEATURE\n" +
				"        issueKey: AAA-000\n" +
				"        issueSummary: summary\n" +
				"        issueType: type\n" +
				"        issueStatus: status\n" +
				"        commitDetail:\n" +
				"          parsedKey: AAA-000\n" +
				"          parsedIssueTracker: JIRA\n" +
				"  bugs: {}\n" +
				"  knownIssues: {}\n" +
				"  breakingChange: {}\n" +
				"  gitRepos:\n" +
				"    - label: label1\n" +
				"      gitRepoID: repoID\n" +
				"      previousVersion: 0.0.0\n" +
				"      version: 1.0.0\n" +
				"      jiraProject: project\n" +
				"      jiraComponent: component\n" +
				"      extractedKeys:\n" +
				"        - parsedKey: AAA-000\n" +
				"          parsedIssueTracker: JIRA\n" +
				"      hasNewFeature: true\n",
			expectedError: nil,
		},
		{
			name:          "Test EnrichWithIssueTrackers with bug",
			parsedCommits: []entities.ParsedCommit{{ParsedKey: "AAA-000", ParsedIssueTracker: "JIRA", IsBreakingChange: false}},
			issues: []entities.Issue{{
				IssueKey:     "AAA-000",
				IssueSummary: "summary",
				IssueType:    "type",
				IssueStatus:  "status",
				Category:     entities.FIXED_BUG,
			}},
			knownIssues: []entities.Issue{},
			expectedContent: "services:\n" +
				"  - label: label1\n" +
				"    gitRepoID: repoID\n" +
				"    previousVersion: 0.0.0\n" +
				"    version: 1.0.0\n" +
				"    jiraProject: project\n" +
				"    jiraComponent: component\n" +
				"generatedValues:\n" +
				"  features: {}\n" +
				"  bugs:\n" +
				"    label1:\n" +
				"      - issueTracker: JIRA\n" +
				"        category: FIXED_BUG\n" +
				"        issueKey: AAA-000\n" +
				"        issueSummary: summary\n" +
				"        issueType: type\n" +
				"        issueStatus: status\n" +
				"        commitDetail:\n" +
				"          parsedKey: AAA-000\n" +
				"          parsedIssueTracker: JIRA\n" +
				"  knownIssues: {}\n" +
				"  breakingChange: {}\n" +
				"  gitRepos:\n" +
				"    - label: label1\n" +
				"      gitRepoID: repoID\n" +
				"      previousVersion: 0.0.0\n" +
				"      version: 1.0.0\n" +
				"      jiraProject: project\n" +
				"      jiraComponent: component\n" +
				"      extractedKeys:\n" +
				"        - parsedKey: AAA-000\n" +
				"          parsedIssueTracker: JIRA\n" +
				"      hasBugFixed: true\n",
			expectedError: nil,
		},
		{
			name:   "Test EnrichWithIssueTrackers known issue",
			issues: []entities.Issue{},
			knownIssues: []entities.Issue{{
				IssueKey:     "AAA-000",
				IssueSummary: "summary",
				IssueType:    "type",
				IssueStatus:  "status",
				Category:     entities.OTHER,
			}},
			expectedContent: "services:\n" +
				"  - label: label1\n" +
				"    gitRepoID: repoID\n" +
				"    previousVersion: 0.0.0\n" +
				"    version: 1.0.0\n" +
				"    jiraProject: project\n" +
				"    jiraComponent: component\n" +
				"generatedValues:\n" +
				"  features: {}\n" +
				"  bugs: {}\n" +
				"  knownIssues:\n" +
				"    label1:\n" +
				"      - issueTracker: JIRA\n" +
				"        category: OTHER\n" +
				"        issueKey: AAA-000\n" +
				"        issueSummary: summary\n" +
				"        issueType: type\n" +
				"        issueStatus: status\n" +
				"  breakingChange: {}\n" +
				"  gitRepos:\n" +
				"    - label: label1\n" +
				"      gitRepoID: repoID\n" +
				"      previousVersion: 0.0.0\n" +
				"      version: 1.0.0\n" +
				"      jiraProject: project\n" +
				"      jiraComponent: component\n",
			expectedError: nil,
		},
		{
			name:          "Test EnrichWithIssueTrackers with breaking change",
			parsedCommits: []entities.ParsedCommit{{ParsedKey: "AAA-000", ParsedIssueTracker: "JIRA", IsBreakingChange: true}},
			issues: []entities.Issue{{
				IssueKey:     "AAA-000",
				IssueSummary: "summary",
				IssueType:    "type",
				IssueStatus:  "status",
				Category:     entities.CLOSED_FEATURE,
			}},
			knownIssues: []entities.Issue{},
			expectedContent: "services:\n" +
				"  - label: label1\n" +
				"    gitRepoID: repoID\n" +
				"    previousVersion: 0.0.0\n" +
				"    version: 1.0.0\n" +
				"    jiraProject: project\n" +
				"    jiraComponent: component\n" +
				"generatedValues:\n" +
				"  features:\n" +
				"    label1:\n" +
				"      - issueTracker: JIRA\n" +
				"        category: CLOSED_FEATURE\n" +
				"        issueKey: AAA-000\n" +
				"        issueSummary: summary\n" +
				"        issueType: type\n" +
				"        issueStatus: status\n" +
				"        commitDetail:\n" +
				"          parsedKey: AAA-000\n" +
				"          parsedIssueTracker: JIRA\n" +
				"          isBreakingChange: true\n" +
				"  bugs: {}\n" +
				"  knownIssues: {}\n" +
				"  breakingChange:\n" +
				"    label1:\n" +
				"      - issueTracker: JIRA\n" +
				"        category: CLOSED_FEATURE\n" +
				"        issueKey: AAA-000\n" +
				"        issueSummary: summary\n" +
				"        issueType: type\n" +
				"        issueStatus: status\n" +
				"        commitDetail:\n" +
				"          parsedKey: AAA-000\n" +
				"          parsedIssueTracker: JIRA\n" +
				"          isBreakingChange: true\n" +
				"  gitRepos:\n" +
				"    - label: label1\n" +
				"      gitRepoID: repoID\n" +
				"      previousVersion: 0.0.0\n" +
				"      version: 1.0.0\n" +
				"      jiraProject: project\n" +
				"      jiraComponent: component\n" +
				"      extractedKeys:\n" +
				"        - parsedKey: AAA-000\n" +
				"          parsedIssueTracker: JIRA\n" +
				"          isBreakingChange: true\n" +
				"      hasBreaking: true\n" +
				"      hasNewFeature: true\n",
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockedCommtis := []entities.Commit{{ID: "commit1", ShortID: "shortID1", Message: "[AAA-1234] With reference"}}

			mockGitLabClient := new(MockGitLabClient)
			mockGitLabClient.On("GetCommits", "repoID", "0.0.0", "1.0.0").Return(mockedCommtis, nil)

			mockRepoParser := new(MockRepoParser)
			mockRepoParser.On("Parse", mockedCommtis).Return(tt.parsedCommits, nil)

			mockIssueTracker := new(MockIssueTracker)
			mockIssueTracker.On("GetIssues", mock.Anything).Return(tt.issues, nil)
			mockIssueTracker.On("GetKnownIssues", mock.Anything).Return(tt.knownIssues, nil)

			values := []byte("" +
				"services:\n" +
				"  - label: label1\n" +
				"    gitRepoID: repoID\n" +
				"    previousVersion: 0.0.0\n" +
				"    version: 1.0.0\n" +
				"    jiraProject: project\n" +
				"    jiraComponent: component\n")

			m, err := model.New(values,
				model.WithRepoClient(mockGitLabClient),
				model.WithRepoParser(mockRepoParser),
				model.WithIssueTracker("JIRA", mockIssueTracker))
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedError, err)
			err = m.EnrichWithRepos()
			assert.NoError(t, err)

			err = m.EnrichWithIssueTrackers()
			assert.NoError(t, err)

			bytes, err := m.Yaml()
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedContent, string(bytes))
		})
	}
}

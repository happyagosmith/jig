package model_test

import (
	"context"
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

func (m *MockRepoParser) GetParsedRecords(id, from, to, mrTargetBranch string) ([]entities.ParsedRepoRecord, error) {
	args := m.Called(id, from, to, mrTargetBranch)
	return args.Get(0).([]entities.ParsedRepoRecord), args.Error(1)
}

func (m *MockRepoParser) GetReleaseURL(id, version string) (string, error) {
	args := m.Called(id)
	return args.Get(0).(string), args.Error(1)
}

func (m *MockRepoParser) GetRepoURL(id string) (string, error) {
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
		name                string
		repoID              string
		previousVersion     string
		version             string
		mockedCommits       []entities.RepoRecord
		mockedMergeRequests []entities.RepoRecord
		wantContent         string
	}{
		{
			name:                "Test EnrichWithGit",
			repoID:              "repo1",
			previousVersion:     "0.0.0",
			version:             "1.0.0",
			mockedCommits:       []entities.RepoRecord{},
			mockedMergeRequests: []entities.RepoRecord{},
			wantContent: "" +
				"services:\n" +
				"  - label: label1\n" +
				"    gitRepoID: repo1\n" +
				"    previousVersion: 0.0.0\n" +
				"    version: 1.0.0\n" +
				"    customAttributes:\n" +
				"      key1: value1\n" +
				"      key2: value2\n" +
				"generatedValues:\n" +
				"  features: {}\n" +
				"  bugs: {}\n" +
				"  knownIssues: {}\n" +
				"  breakingChange: {}\n" +
				"  gitRepos:\n" +
				"    - label: label1\n" +
				"      gitRepoID: repo1\n" +
				"      previousVersion: 0.0.0\n" +
				"      version: 1.0.0\n" +
				"      customAttributes:\n" +
				"        key1: value1\n" +
				"        key2: value2\n", 
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepoParser := new(MockRepoParser)
			mockRepoParser.On("GetParsedRecords", tt.repoID, tt.previousVersion, tt.version, "").Return([]entities.ParsedRepoRecord{}, nil)

			values := []byte(fmt.Sprintf(`
services:
  - label: label1
    gitRepoID: %s
    previousVersion: %s
    version: %s
    customAttributes:
      key1: value1
      key2: value2
`, tt.repoID, tt.previousVersion, tt.version))

			m, err := model.New(values,
				model.WithRepoService(mockRepoParser))
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			err = m.EnrichWithRepos()
			assert.NoError(t, err)

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

func (m *MockIssueTracker) GetIssues(_ context.Context, _ *entities.EnrichedRepo, keys []string) ([]entities.Issue, error) {
	args := m.Called(keys)
	return args.Get(0).([]entities.Issue), args.Error(1)
}

func (m *MockIssueTracker) GetKnownIssues(_ context.Context, repo *entities.EnrichedRepo) ([]entities.Issue, error) {
	args := m.Called(repo)
	return args.Get(0).([]entities.Issue), args.Error(1)
}

func TestEnrichWithIssueTrackers(t *testing.T) {
	tests := []struct {
		name               string
		mockedParsedRecord []entities.ParsedRepoRecord
		issues             []entities.Issue
		knownIssues        []entities.Issue
		expectedContent    string
		expectedError      error
	}{
		{
			name:               "Test EnrichWithIssueTrackers no issues",
			mockedParsedRecord: []entities.ParsedRepoRecord{{ParsedKey: "AAA-000", ParsedIssueTracker: "JIRA"}},
			issues:             []entities.Issue{},
			knownIssues:        []entities.Issue{},
			expectedContent: "services:\n" +
				"  - label: label1\n" +
				"    gitRepoID: repoID\n" +
				"    previousVersion: 0.0.0\n" +
				"    version: 1.0.0\n" +
				"    jiraProject: project\n" +
				"    jiraComponent: component\n" +
				"    customAttributes:\n" +
				"      key1: value1\n" +
				"      key2: value2\n" +
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
				"      customAttributes:\n" +
				"        key1: value1\n" +
				"        key2: value2\n" +
				"      extractedKeys:\n" +
				"        - parsedKey: AAA-000\n" +
				"          parsedIssueTracker: JIRA\n",
			expectedError: nil,
		},
		{
			name:               "Test EnrichWithIssueTrackers with features",
			mockedParsedRecord: []entities.ParsedRepoRecord{{ParsedKey: "AAA-000", ParsedIssueTracker: "JIRA", IsBreakingChange: false}},
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
				"    customAttributes:\n" +
				"      key1: value1\n" +
				"      key2: value2\n" +
				"generatedValues:\n" +
				"  features:\n" +
				"    label1:\n" +
				"      - issueTracker: JIRA\n" +
				"        issueKey: AAA-000\n" +
				"        issueSummary: summary\n" +
				"        issueCategory: CLOSED_FEATURE\n" +
				"        issueDetail:\n" +
				"          extractedCategory: CLOSED_FEATURE\n" +
				"          issueKey: AAA-000\n" +
				"          issueSummary: summary\n" +
				"          issueType: type\n" +
				"          issueStatus: status\n" +
				"        repoDetail:\n" +
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
				"      customAttributes:\n" +
				"        key1: value1\n" +
				"        key2: value2\n" +
				"      extractedKeys:\n" +
				"        - parsedKey: AAA-000\n" +
				"          parsedIssueTracker: JIRA\n" +
				"      hasNewFeature: true\n",
			expectedError: nil,
		},
		{
			name:               "Test EnrichWithIssueTrackers with bug",
			mockedParsedRecord: []entities.ParsedRepoRecord{{ParsedKey: "AAA-000", ParsedIssueTracker: "JIRA", IsBreakingChange: false}},
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
				"    customAttributes:\n" +
				"      key1: value1\n" +
				"      key2: value2\n" +
				"generatedValues:\n" +
				"  features: {}\n" +
				"  bugs:\n" +
				"    label1:\n" +
				"      - issueTracker: JIRA\n" +
				"        issueKey: AAA-000\n" +
				"        issueSummary: summary\n" +
				"        issueCategory: FIXED_BUG\n" +
				"        issueDetail:\n" +
				"          extractedCategory: FIXED_BUG\n" +
				"          issueKey: AAA-000\n" +
				"          issueSummary: summary\n" +
				"          issueType: type\n" +
				"          issueStatus: status\n" +
				"        repoDetail:\n" +
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
				"      customAttributes:\n" +
				"        key1: value1\n" +
				"        key2: value2\n" +
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
				"    customAttributes:\n" +
				"      key1: value1\n" +
				"      key2: value2\n" +
				"generatedValues:\n" +
				"  features: {}\n" +
				"  bugs: {}\n" +
				"  knownIssues:\n" +
				"    label1:\n" +
				"      - issueTracker: JIRA\n" +
				"        issueKey: AAA-000\n" +
				"        issueSummary: summary\n" +
				"        issueCategory: OTHER\n" +
				"        issueDetail:\n" +
				"          extractedCategory: OTHER\n" +
				"          issueKey: AAA-000\n" +
				"          issueSummary: summary\n" +
				"          issueType: type\n" +
				"          issueStatus: status\n" +
				"  breakingChange: {}\n" +
				"  gitRepos:\n" +
				"    - label: label1\n" +
				"      gitRepoID: repoID\n" +
				"      previousVersion: 0.0.0\n" +
				"      version: 1.0.0\n" +
				"      jiraProject: project\n" +
				"      jiraComponent: component\n" +
				"      customAttributes:\n" +
				"        key1: value1\n" +
				"        key2: value2\n",
			expectedError: nil,
		},
		{
			name:               "Test EnrichWithIssueTrackers with breaking change",
			mockedParsedRecord: []entities.ParsedRepoRecord{{ParsedKey: "AAA-000", ParsedIssueTracker: "JIRA", IsBreakingChange: true}},
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
				"    customAttributes:\n" +
				"      key1: value1\n" +
				"      key2: value2\n" +
				"generatedValues:\n" +
				"  features:\n" +
				"    label1:\n" +
				"      - issueTracker: JIRA\n" +
				"        issueKey: AAA-000\n" +
				"        issueSummary: summary\n" +
				"        issueCategory: CLOSED_FEATURE\n" +
				"        issueDetail:\n" +
				"          extractedCategory: CLOSED_FEATURE\n" +
				"          issueKey: AAA-000\n" +
				"          issueSummary: summary\n" +
				"          issueType: type\n" +
				"          issueStatus: status\n" +
				"        repoDetail:\n" +
				"          parsedKey: AAA-000\n" +
				"          parsedIssueTracker: JIRA\n" +
				"          isBreakingChange: true\n" +
				"  bugs: {}\n" +
				"  knownIssues: {}\n" +
				"  breakingChange:\n" +
				"    label1:\n" +
				"      - issueTracker: JIRA\n" +
				"        issueKey: AAA-000\n" +
				"        issueSummary: summary\n" +
				"        issueCategory: CLOSED_FEATURE\n" +
				"        issueDetail:\n" +
				"          extractedCategory: CLOSED_FEATURE\n" +
				"          issueKey: AAA-000\n" +
				"          issueSummary: summary\n" +
				"          issueType: type\n" +
				"          issueStatus: status\n" +
				"        repoDetail:\n" +
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
				"      customAttributes:\n" +
				"        key1: value1\n" +
				"        key2: value2\n" +
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
			mockRepoParser := new(MockRepoParser)
			mockRepoParser.On("GetParsedRecords", "repoID", "0.0.0", "1.0.0", "").Return(tt.mockedParsedRecord, nil)

			mockIssueTracker := new(MockIssueTracker)
			mockIssueTracker.On("GetIssues", []string{"AAA-000"}).Return(tt.issues, nil)
			mockIssueTracker.On("GetKnownIssues", mock.Anything).Return(tt.knownIssues, nil)

			values := []byte("" +
				"services:\n" +
				"  - label: label1\n" +
				"    gitRepoID: repoID\n" +
				"    previousVersion: 0.0.0\n" +
				"    version: 1.0.0\n" +
				"    jiraProject: project\n" +
				"    jiraComponent: component\n" +
				"    customAttributes:\n" +
				"      key1: value1\n" +
				"      key2: value2\n" )

			m, err := model.New(values,
				model.WithRepoService(mockRepoParser),
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

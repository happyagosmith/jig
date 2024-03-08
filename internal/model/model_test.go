package model_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/happyagosmith/jig/internal/model"
	git "github.com/happyagosmith/jig/internal/repositories"
	"github.com/happyagosmith/jig/internal/trackers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockGitLabClient struct {
	mock.Mock
}

func (m *MockGitLabClient) ExtractCommits(repoID, fromTag, toTag string) ([]git.CommitDetail, error) {
	args := m.Called(repoID, fromTag, toTag)
	return args.Get(0).([]git.CommitDetail), args.Error(1)
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

	err = m.SetVersions(os.TempDir())
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
			mockGitLabClient := new(MockGitLabClient)
			mockGitLabClient.On("ExtractCommits", tt.repoID, tt.previousVersion, tt.version).Return([]git.CommitDetail{}, nil)

			values := []byte(fmt.Sprintf(`
services:
  - label: label1
    gitRepoID: %s
    previousVersion: %s
    version: %s
`, tt.repoID, tt.previousVersion, tt.version))

			m, err := model.New(values, model.WithRepoSRV(mockGitLabClient))
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			err = m.EnrichWithGit()

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

func (m *MockIssueTracker) GetIssues(keys []string) ([]trackers.IssueDetail, error) {
	args := m.Called(keys)
	return args.Get(0).([]trackers.IssueDetail), args.Error(1)
}

func (m *MockIssueTracker) GetKnownIssues(project, component string) ([]trackers.IssueDetail, error) {
	args := m.Called(project, component)
	return args.Get(0).([]trackers.IssueDetail), args.Error(1)
}

func TestEnrichWithIssueTrackers(t *testing.T) {
	tests := []struct {
		name            string
		commits         []git.CommitDetail
		issues          []trackers.IssueDetail
		knownIssues     []trackers.IssueDetail
		expectedContent string
		expectedError   error
	}{
		{
			name:        "Test EnrichWithIssueTrackers no issues",
			commits:     []git.CommitDetail{{ParsedKey: "AAA-000", ParsedIssueTracker: "JIRA"}},
			issues:      []trackers.IssueDetail{},
			knownIssues: []trackers.IssueDetail{},
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
			name:    "Test EnrichWithIssueTrackers with features",
			commits: []git.CommitDetail{{ParsedKey: "AAA-000", ParsedIssueTracker: "JIRA", IsBreakingChange: false}},
			issues: []trackers.IssueDetail{{
				IssueKey:     "AAA-000",
				IssueSummary: "summary",
				IssueType:    "type",
				IssueStatus:  "status",
				Category:     trackers.CLOSED_FEATURE,
			},
			},
			knownIssues: []trackers.IssueDetail{},
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
			name:    "Test EnrichWithIssueTrackers with bug",
			commits: []git.CommitDetail{{ParsedKey: "AAA-000", ParsedIssueTracker: "JIRA", IsBreakingChange: false}},
			issues: []trackers.IssueDetail{{
				IssueKey:     "AAA-000",
				IssueSummary: "summary",
				IssueType:    "type",
				IssueStatus:  "status",
				Category:     trackers.FIXED_BUG,
			}},
			knownIssues: []trackers.IssueDetail{},
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
			issues: []trackers.IssueDetail{},
			knownIssues: []trackers.IssueDetail{{
				IssueKey:     "AAA-000",
				IssueSummary: "summary",
				IssueType:    "type",
				IssueStatus:  "status",
				Category:     trackers.OTHER,
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
			name:    "Test EnrichWithIssueTrackers with breaking change",
			commits: []git.CommitDetail{{ParsedKey: "AAA-000", ParsedIssueTracker: "JIRA", IsBreakingChange: true}},
			issues: []trackers.IssueDetail{{
				IssueKey:     "AAA-000",
				IssueSummary: "summary",
				IssueType:    "type",
				IssueStatus:  "status",
				Category:     trackers.CLOSED_FEATURE,
			}},
			knownIssues: []trackers.IssueDetail{},
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
			mockGitLabClient := new(MockGitLabClient)
			mockIssueTracker := new(MockIssueTracker)

			mockGitLabClient.On("ExtractCommits", "repoID", "0.0.0", "1.0.0").Return(tt.commits, nil)
			mockIssueTracker.On("GetIssues", mock.Anything).Return(tt.issues, nil)
			mockIssueTracker.On("GetKnownIssues", "project", "component").Return(tt.knownIssues, nil)

			values := []byte("" +
				"services:\n" +
				"  - label: label1\n" +
				"    gitRepoID: repoID\n" +
				"    previousVersion: 0.0.0\n" +
				"    version: 1.0.0\n" +
				"    jiraProject: project\n" +
				"    jiraComponent: component\n")
			m, err := model.New(values, model.WithRepoSRV(mockGitLabClient), model.WithIssueTracker("JIRA", mockIssueTracker))
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedError, err)
			err = m.EnrichWithGit()
			assert.NoError(t, err)

			err = m.EnrichWithIssueTrackers()
			assert.NoError(t, err)

			bytes, err := m.Yaml()
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedContent, string(bytes))
		})
	}
}

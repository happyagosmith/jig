package parsers_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/happyagosmith/jig/internal/entities"
	"github.com/happyagosmith/jig/internal/parsers"
	"github.com/happyagosmith/jig/internal/repoclients"
	"github.com/stretchr/testify/assert"
)

func TestGitLabCommitParse(t *testing.T) {
	type want struct {
		idx int
		cd  entities.ParsedCommit
	}
	tests := []struct {
		name               string
		mockGitLabResponse string
		issuePatterns      []parsers.IssuePattern
		wantResultLen      int
		wantCommitDetails  []want
		keepCCWithoutScope bool
	}{
		{
			name:               "parse jira commits",
			mockGitLabResponse: "testdata/git-compare.json",
			issuePatterns: []parsers.IssuePattern{
				{IssueTracker: "jira", Pattern: `\w+-\d+`}},
			wantResultLen: 2,
			wantCommitDetails: []want{
				{
					idx: 0,
					cd: entities.ParsedCommit{
						ParsedKey:          "AAA-1234",
						ParsedIssueTracker: "JIRA",
						IsBreakingChange:   false,
						ParsedCategory:     entities.UNKNOWN,
						Summary:            "With reference",
						CommitID:           "commit1",
						Message:            "[AAA-1234] With reference\n",
					},
				},
				{
					idx: 1,
					cd: entities.ParsedCommit{
						ParsedKey:          "AAA-5678",
						ParsedIssueTracker: "JIRA",
						IsBreakingChange:   false,
						ParsedCategory:     entities.UNKNOWN,
						Summary:            "Different reference",
						CommitID:           "commit3",
						Message:            "[AAA-5678] Different reference\n",
					},
				},
			},
		},
		{
			name:               "parse feature from conventional commit",
			mockGitLabResponse: "testdata/git-compare-conventional-commit.json",
			issuePatterns: []parsers.IssuePattern{
				{IssueTracker: "jira", Pattern: `j_(.+)`}},
			wantResultLen: 4,
			wantCommitDetails: []want{
				{
					idx: 0,
					cd: entities.ParsedCommit{
						ParsedKey:          "CC-123",
						ParsedIssueTracker: "JIRA",
						IsBreakingChange:   false,
						ParsedCategory:     entities.FEATURE,
						Summary:            "this is a feature tracked in jira",
						CommitID:           "commit1",
						Message:            "feat(j_CC-123): this is a feature tracked in jira",
					},
				},
			},
		},
		{
			name:               "parse bug fixed from conventional commit",
			mockGitLabResponse: "testdata/git-compare-conventional-commit.json",
			issuePatterns: []parsers.IssuePattern{
				{IssueTracker: "jira", Pattern: `j_(.+)`}},
			wantResultLen: 4,
			wantCommitDetails: []want{
				{
					idx: 1,
					cd: entities.ParsedCommit{
						ParsedKey:          "CC-456",
						ParsedIssueTracker: "JIRA",
						IsBreakingChange:   false,
						ParsedCategory:     entities.BUG_FIX,
						Summary:            "this is a bug fixed tracked in jira",
						CommitID:           "commit2",
						Message:            "fix(j_CC-456): this is a bug fixed tracked in jira",
					},
				},
			},
		},
		{
			name:               "parse breaking change from conventional commit",
			mockGitLabResponse: "testdata/git-compare-conventional-commit.json",
			issuePatterns: []parsers.IssuePattern{
				{IssueTracker: "jira", Pattern: `j_(.+)`}},
			wantResultLen: 4,
			wantCommitDetails: []want{
				{
					idx: 2,
					cd: entities.ParsedCommit{
						ParsedKey:          "CC-789",
						ParsedIssueTracker: "JIRA",
						IsBreakingChange:   true,
						ParsedCategory:     entities.BUG_FIX,
						Summary:            "this is a breaking change tracked in jira",
						CommitID:           "commit3",
						Message:            "fix(j_CC-789)!: this is a breaking change tracked in jira",
					},
				},
			},
		},
		{
			name:               "parse unknown issue tracker from conventional commit",
			mockGitLabResponse: "testdata/git-compare-conventional-commit.json",
			issuePatterns: []parsers.IssuePattern{
				{IssueTracker: "jira", Pattern: `j_(.+)`}},
			wantResultLen: 4,
			wantCommitDetails: []want{
				{
					idx: 3,
					cd: entities.ParsedCommit{
						ParsedKey:          "CC-222",
						ParsedIssueTracker: "NONE",
						IsBreakingChange:   false,
						ParsedCategory:     entities.BUG_FIX,
						Summary:            "this has an unknown issue tracker",
						CommitID:           "commit4",
						Message:            "fix(CC-222): this has an unknown issue tracker",
					},
				},
			},
		},
		{
			name:               "add feature without issuekey from conventional commit",
			mockGitLabResponse: "testdata/git-compare-conventional-commit.json",
			issuePatterns: []parsers.IssuePattern{
				{IssueTracker: "jira", Pattern: `j_(.+)`}},
			keepCCWithoutScope: true,
			wantResultLen:      5,
			wantCommitDetails: []want{
				{
					idx: 4,
					cd: entities.ParsedCommit{
						ParsedKey:          "",
						ParsedIssueTracker: "NONE",
						IsBreakingChange:   false,
						ParsedCategory:     entities.BUG_FIX,
						Summary:            "this has an unknown issue tracker and no issueKey",
						CommitID:           "commit5",
						Message:            "fix: this has an unknown issue tracker and no issueKey",
					},
				},
			},
		},
		{
			name:               "not add feature without issuekey from conventional commit",
			mockGitLabResponse: "testdata/git-compare-conventional-commit.json",
			issuePatterns: []parsers.IssuePattern{
				{IssueTracker: "jira", Pattern: `j_(.+)`}},
			wantResultLen:     4,
			wantCommitDetails: []want{},
		},
		{
			name:               "add commit from close pattern",
			mockGitLabResponse: "testdata/git-compare-close-pattern.json",
			issuePatterns: []parsers.IssuePattern{
				{IssueTracker: "git", Pattern: "#([A-Z0-9]+)"}},
			wantResultLen: 2,
			wantCommitDetails: []want{
				{
					idx: 0,
					cd: entities.ParsedCommit{
						ParsedKey:          "1",
						ParsedIssueTracker: "GIT",
						IsBreakingChange:   false,
						ParsedCategory:     entities.FEATURE,
						Summary:            "ci: Update README.md",
						CommitID:           "commit0",
						Message:            "ci: Update README.md\n\nthis commit is to show if it works closes #1",
					},
				},
				{
					idx: 1,
					cd: entities.ParsedCommit{
						ParsedKey:          "2",
						ParsedIssueTracker: "GIT",
						IsBreakingChange:   false,
						ParsedCategory:     entities.FEATURE,
						Summary:            "Merge branch '2-this-is-an-issue-to-test-the-mr' into 'main'",
						CommitID:           "commit2",
						Message:            "Merge branch '2-this-is-an-issue-to-test-the-mr' into 'main'\n\nResolve \"this is an issue to test the MR\"\n\nCloses #2\n\nSee merge request demo!1",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)
				b, _ := os.ReadFile(tt.mockGitLabResponse)
				w.Write(b)
			}))
			defer srv.Close()

			gc, err := repoclients.NewGitLab(srv.URL, "token")
			assert.NoError(t, err, "NewGit error must be nil")
			commits, err := gc.GetCommits("123", "from", "to")
			assert.NoError(t, err, "GetCommits error must be nil")

			gp, err := parsers.New(tt.issuePatterns,
				parsers.WithCustomPattern(`\[(?P<scope>[^\]]*)\](?P<subject>.*)`),
				parsers.WithKeepCCWithoutScope(tt.keepCCWithoutScope))
			assert.NoError(t, err, "NewGit error must be nil")
			gotCds, err := gp.Parse(commits)

			assert.NoError(t, err, "Parse error must be nil")
			assert.Equal(t, tt.wantResultLen, len(gotCds))

			for _, want := range tt.wantCommitDetails {
				idx := want.idx
				assert.Equal(t, want.cd.ParsedKey, gotCds[idx].ParsedKey)
				assert.Equal(t, want.cd.ParsedIssueTracker, gotCds[idx].ParsedIssueTracker)
				assert.Equal(t, want.cd.IsBreakingChange, gotCds[idx].IsBreakingChange)
				assert.Equal(t, want.cd.ParsedCategory, gotCds[idx].ParsedCategory)
				assert.Equal(t, want.cd.Summary, gotCds[idx].Summary)
				assert.Equal(t, want.cd.CommitID, gotCds[idx].CommitID)
				assert.Equal(t, want.cd.Message, gotCds[idx].Message)
			}
		})
	}
}

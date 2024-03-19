package repo_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/happyagosmith/jig/internal/entities"
	"github.com/happyagosmith/jig/internal/parsers"
	"github.com/happyagosmith/jig/internal/repo"
	"github.com/happyagosmith/jig/internal/repo/clients"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfiguration(t *testing.T) {
	tests := []struct {
		name            string
		defaultMRBranch string
		wantMRBranch    string
		argMRBranch     string
	}{
		{
			name:            "MR default branch",
			defaultMRBranch: "dfMRBranch",
			wantMRBranch:    "dfMRBranch",
			argMRBranch:     "",
		},
		{
			name:            "MR set branch",
			defaultMRBranch: "dfMRBranch",
			wantMRBranch:    "main",
			argMRBranch:     "main",
		},
		{
			name:            "MR set branch",
			defaultMRBranch: "",
			wantMRBranch:    "main",
			argMRBranch:     "main",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gitSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				params := r.URL.Query()
				if r.URL.Path == "/api/v4/projects/123/repository/compare" {
					w.WriteHeader(200)
					assertParam(t, "from", "from", params)
					assertParam(t, "to", "to", params)
					w.Write([]byte("{\"commits\": [{\"id\": \"commit0\", \"message\": \"No reference\"}]}"))
				} else if r.URL.Path == "/api/v4/projects/123/merge_requests" {
					w.WriteHeader(200)
					assertParam(t, "state", "merged", params)
					assertParam(t, "target_branch", tt.wantMRBranch, params)
					w.Write([]byte("[]"))
				} else {
					http.Error(w, "Not found", http.StatusNotFound)
				}
			}))
			defer gitSrv.Close()

			gc, err := clients.NewGitLab(gitSrv.URL, "token")
			assert.NoError(t, err, "NewGit error must be nil")

			gp, err := repo.New(gc, []parsers.IssuePattern{},
				repo.WithDefaultMRBranch(tt.defaultMRBranch))
			assert.NoError(t, err, "NewGit error must be nil")
			_, err = gp.GetParsedRecords("123", "from", "to", tt.argMRBranch)
			assert.NoError(t, err, "Parse error must be nil")
		})
	}
}

func TestGitLabCommitParse(t *testing.T) {
	type want struct {
		idx int
		cd  entities.ParsedRepoRecord
	}
	tests := []struct {
		name                      string
		mockGitLabCompareResponse string
		issuePatterns             []parsers.IssuePattern
		wantResultLen             int
		wantParsedRepoRecords     []want
		keepCCWithoutScope        bool
	}{
		{
			name:                      "parse jira commits",
			mockGitLabCompareResponse: "testdata/git-compare.json",
			issuePatterns: []parsers.IssuePattern{
				{IssueTracker: "jira", Pattern: `\w+-\d+`}},
			wantResultLen: 2,
			wantParsedRepoRecords: []want{
				{
					idx: 0,
					cd: entities.ParsedRepoRecord{
						ParsedKey:          "AAA-1234",
						ParsedIssueTracker: "JIRA",
						IsBreakingChange:   false,
						ParsedCategory:     entities.UNKNOWN,
						ParsedSummary:      "With reference",
						Parser:             "customParser",
						RepoRecord:         generalCommitRepoRecord(1, "[AAA-1234] With reference\n"),
					},
				},
				{
					idx: 1,
					cd: entities.ParsedRepoRecord{
						ParsedKey:          "AAA-5678",
						ParsedIssueTracker: "JIRA",
						IsBreakingChange:   false,
						ParsedCategory:     entities.UNKNOWN,
						ParsedSummary:      "Different reference",
						Parser:             "customParser",
						RepoRecord:         generalCommitRepoRecord(3, "[AAA-5678] Different reference\n"),
					},
				},
			},
		},
		{
			name:                      "parse feature from conventional commit",
			mockGitLabCompareResponse: "testdata/git-compare-conventional-commit.json",
			issuePatterns: []parsers.IssuePattern{
				{IssueTracker: "jira", Pattern: `j_(.+)`}},
			wantResultLen: 4,
			wantParsedRepoRecords: []want{
				{
					idx: 0,
					cd: entities.ParsedRepoRecord{
						ParsedKey:          "CC-123",
						ParsedIssueTracker: "JIRA",
						IsBreakingChange:   false,
						ParsedCategory:     entities.FEATURE,
						ParsedSummary:      "this is a feature tracked in jira",
						Parser:             "conventionalParser",
						ParsedType:         "feat",
						RepoRecord:         generalCommitRepoRecord(1, "feat(j_CC-123): this is a feature tracked in jira"),
					},
				},
			},
		},
		{
			name:                      "parse bug fixed from conventional commit",
			mockGitLabCompareResponse: "testdata/git-compare-conventional-commit.json",
			issuePatterns: []parsers.IssuePattern{
				{IssueTracker: "jira", Pattern: `j_(.+)`}},
			wantResultLen: 4,
			wantParsedRepoRecords: []want{
				{
					idx: 1,
					cd: entities.ParsedRepoRecord{
						ParsedKey:          "CC-456",
						ParsedIssueTracker: "JIRA",
						IsBreakingChange:   false,
						ParsedCategory:     entities.BUG_FIX,
						ParsedSummary:      "this is a bug fixed tracked in jira",
						Parser:             "conventionalParser",
						ParsedType:         "fix",
						RepoRecord:         generalCommitRepoRecord(2, "fix(j_CC-456): this is a bug fixed tracked in jira"),
					},
				},
			},
		},
		{
			name:                      "parse breaking change from conventional commit",
			mockGitLabCompareResponse: "testdata/git-compare-conventional-commit.json",
			issuePatterns: []parsers.IssuePattern{
				{IssueTracker: "jira", Pattern: `j_(.+)`}},
			wantResultLen: 4,
			wantParsedRepoRecords: []want{
				{
					idx: 2,
					cd: entities.ParsedRepoRecord{
						ParsedKey:          "CC-789",
						ParsedIssueTracker: "JIRA",
						IsBreakingChange:   true,
						ParsedCategory:     entities.BUG_FIX,
						ParsedSummary:      "this is a breaking change tracked in jira",
						Parser:             "conventionalParser",
						ParsedType:         "fix",
						RepoRecord:         generalCommitRepoRecord(3, "fix(j_CC-789)!: this is a breaking change tracked in jira"),
					},
				},
			},
		},
		{
			name:                      "parse unknown issue tracker from conventional commit",
			mockGitLabCompareResponse: "testdata/git-compare-conventional-commit.json",
			issuePatterns: []parsers.IssuePattern{
				{IssueTracker: "jira", Pattern: `j_(.+)`}},
			wantResultLen: 4,
			wantParsedRepoRecords: []want{
				{
					idx: 3,
					cd: entities.ParsedRepoRecord{
						ParsedKey:          "CC-222",
						ParsedIssueTracker: "NONE",
						IsBreakingChange:   false,
						ParsedCategory:     entities.BUG_FIX,
						ParsedSummary:      "this has an unknown issue tracker",
						Parser:             "conventionalParser",
						ParsedType:         "fix",
						RepoRecord:         generalCommitRepoRecord(4, "fix(CC-222): this has an unknown issue tracker"),
					},
				},
			},
		},
		{
			name:                      "add feature without issuekey from conventional commit",
			mockGitLabCompareResponse: "testdata/git-compare-conventional-commit.json",
			issuePatterns: []parsers.IssuePattern{
				{IssueTracker: "jira", Pattern: `j_(.+)`}},
			keepCCWithoutScope: true,
			wantResultLen:      5,
			wantParsedRepoRecords: []want{
				{
					idx: 4,
					cd: entities.ParsedRepoRecord{
						ParsedKey:          "",
						ParsedIssueTracker: "NONE",
						IsBreakingChange:   false,
						ParsedCategory:     entities.BUG_FIX,
						ParsedSummary:      "this has an unknown issue tracker and no issueKey",
						Parser:             "conventionalParser",
						ParsedType:         "fix",
						RepoRecord:         generalCommitRepoRecord(5, "fix: this has an unknown issue tracker and no issueKey"),
					},
				},
			},
		},
		{
			name:                      "not add feature without issuekey from conventional commit",
			mockGitLabCompareResponse: "testdata/git-compare-conventional-commit.json",
			issuePatterns: []parsers.IssuePattern{
				{IssueTracker: "jira", Pattern: `j_(.+)`}},
			wantResultLen:         4,
			wantParsedRepoRecords: []want{},
		},
		{
			name:                      "add commit from close pattern",
			mockGitLabCompareResponse: "testdata/git-compare-close-pattern.json",
			issuePatterns: []parsers.IssuePattern{
				{IssueTracker: "git", Pattern: "#([A-Z0-9]+)"}},
			wantResultLen: 2,
			wantParsedRepoRecords: []want{
				{
					idx: 0,
					cd: entities.ParsedRepoRecord{
						ParsedKey:          "1",
						ParsedIssueTracker: "GIT",
						IsBreakingChange:   false,
						ParsedCategory:     entities.FEATURE,
						ParsedSummary:      "",
						Parser:             "closingPattern",
						ParsedType:         "close",
						RepoRecord:         generalCommitRepoRecord(0, "ci: Update README.md\n\nthis commit is to show if it works closes #1"),
					},
				},
				{
					idx: 1,
					cd: entities.ParsedRepoRecord{
						ParsedKey:          "2",
						ParsedIssueTracker: "GIT",
						IsBreakingChange:   false,
						ParsedCategory:     entities.FEATURE,
						ParsedSummary:      "",
						Parser:             "closingPattern",
						ParsedType:         "close",
						RepoRecord:         generalCommitRepoRecord(2, "Merge branch '2-this-is-an-issue-to-test-the-mr' into 'main'\n\nResolve \"this is an issue to test the MR\"\n\nCloses #2\n\nSee merge request demo!1"),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gitsrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/api/v4/projects/123/repository/compare" {
					w.WriteHeader(200)
					b, err := os.ReadFile(tt.mockGitLabCompareResponse)
					if err != nil {
						t.Fatal(err)
					}
					w.Write(b)
				} else if r.URL.Path == "/api/v4/projects/123/merge_requests" {
					w.WriteHeader(200)
					w.Write([]byte("[]"))
				} else {
					http.Error(w, "Not found", http.StatusNotFound)
				}
			}))
			defer gitsrv.Close()

			gc, err := clients.NewGitLab(gitsrv.URL, "token")
			assert.NoError(t, err, "NewGit error must be nil")

			gp, err := repo.New(gc, tt.issuePatterns,
				repo.WithCustomPattern(`\[(?P<scope>[^\]]*)\](?P<subject>.*)`),
				repo.WithKeepCCWithoutScope(tt.keepCCWithoutScope))
			assert.NoError(t, err, "NewGit error must be nil")
			gotCds, err := gp.GetParsedRecords("123", "from", "to", "")

			assert.NoError(t, err, "Parse error must be nil")
			assert.Equal(t, tt.wantResultLen, len(gotCds))

			for _, want := range tt.wantParsedRepoRecords {
				idx := want.idx
				assert.Equal(t, want.cd, gotCds[idx])
			}
		})
	}
}

func TestGitLabMRParse(t *testing.T) {
	type want struct {
		idx int
		cd  entities.ParsedRepoRecord
	}
	tests := []struct {
		name                  string
		mrDescription         string
		issuePatterns         []parsers.IssuePattern
		wantParsedRepoRecords []want
		keepCCWithoutScope    bool
	}{
		{
			name:          "parse jira commits",
			mrDescription: "[AAA-1234] This is a test MR",
			issuePatterns: []parsers.IssuePattern{
				{IssueTracker: "jira", Pattern: `\w+-\d+`}},
			wantParsedRepoRecords: []want{
				{
					idx: 0,
					cd: entities.ParsedRepoRecord{
						ParsedKey:          "AAA-1234",
						ParsedIssueTracker: "JIRA",
						IsBreakingChange:   false,
						ParsedCategory:     entities.UNKNOWN,
						ParsedSummary:      "This is a test MR",
						Parser:             "customParser",
					},
				},
			},
		},
		{
			name:          "parse feature from conventional commit",
			mrDescription: "feat(j_AAA-1234): This is a test MR",
			issuePatterns: []parsers.IssuePattern{
				{IssueTracker: "jira", Pattern: `j_(.+)`}},
			wantParsedRepoRecords: []want{
				{
					idx: 0,
					cd: entities.ParsedRepoRecord{
						ParsedKey:          "AAA-1234",
						ParsedIssueTracker: "JIRA",
						IsBreakingChange:   false,
						ParsedCategory:     entities.FEATURE,
						ParsedSummary:      "This is a test MR",
						Parser:             "conventionalParser",
						ParsedType:         "feat",
					},
				},
			},
		},
		{
			name:          "parse bug fixed from conventional commit",
			mrDescription: "fix(j_AAA-1234): This is a test MR",
			issuePatterns: []parsers.IssuePattern{
				{IssueTracker: "jira", Pattern: `j_(.+)`}},
			wantParsedRepoRecords: []want{
				{
					idx: 0,
					cd: entities.ParsedRepoRecord{
						ParsedKey:          "AAA-1234",
						ParsedIssueTracker: "JIRA",
						IsBreakingChange:   false,
						ParsedCategory:     entities.BUG_FIX,
						ParsedSummary:      "This is a test MR",
						Parser:             "conventionalParser",
						ParsedType:         "fix",
					},
				},
			},
		},
		{
			name:          "parse breaking change from conventional commit",
			mrDescription: "fix(j_AAA-1234)!: This is a test MR",
			issuePatterns: []parsers.IssuePattern{
				{IssueTracker: "jira", Pattern: `j_(.+)`}},
			wantParsedRepoRecords: []want{
				{
					idx: 0,
					cd: entities.ParsedRepoRecord{
						ParsedKey:          "AAA-1234",
						ParsedIssueTracker: "JIRA",
						IsBreakingChange:   true,
						ParsedCategory:     entities.BUG_FIX,
						ParsedSummary:      "This is a test MR",
						Parser:             "conventionalParser",
						ParsedType:         "fix",
					},
				},
			},
		},
		{
			name:          "parse unknown issue tracker from conventional commit",
			mrDescription: "fix(UNKNOWN-1234): this has an unknown issue tracker",
			issuePatterns: []parsers.IssuePattern{
				{IssueTracker: "jira", Pattern: `j_(.+)`}},
			wantParsedRepoRecords: []want{
				{
					idx: 0,
					cd: entities.ParsedRepoRecord{
						ParsedKey:          "UNKNOWN-1234",
						ParsedIssueTracker: "NONE",
						IsBreakingChange:   false,
						ParsedCategory:     entities.BUG_FIX,
						ParsedSummary:      "this has an unknown issue tracker",
						Parser:             "conventionalParser",
						ParsedType:         "fix",
					},
				},
			},
		},
		{
			name:          "add feature without issuekey from conventional commit",
			mrDescription: "fix: This is a test MR",
			issuePatterns: []parsers.IssuePattern{
				{IssueTracker: "jira", Pattern: `j_(.+)`}},
			keepCCWithoutScope: true,
			wantParsedRepoRecords: []want{
				{
					idx: 0,
					cd: entities.ParsedRepoRecord{
						ParsedKey:          "",
						ParsedIssueTracker: "NONE",
						IsBreakingChange:   false,
						ParsedCategory:     entities.BUG_FIX,
						ParsedSummary:      "This is a test MR",
						Parser:             "conventionalParser",
						ParsedType:         "fix",
					},
				},
			},
		},
		{
			name:          "not add feature without issuekey from conventional commit",
			mrDescription: "fix: This is a test MR",
			issuePatterns: []parsers.IssuePattern{
				{IssueTracker: "jira", Pattern: `j_(.+)`}},
			wantParsedRepoRecords: []want{},
		},
		{
			name:          "add commit from close pattern",
			mrDescription: "fixes #1",
			issuePatterns: []parsers.IssuePattern{
				{IssueTracker: "git", Pattern: "#([A-Z0-9]+)"}},
			wantParsedRepoRecords: []want{
				{
					idx: 0,
					cd: entities.ParsedRepoRecord{
						ParsedKey:          "1",
						ParsedIssueTracker: "GIT",
						IsBreakingChange:   false,
						ParsedCategory:     entities.BUG_FIX,
						ParsedSummary:      "",
						ParsedType:         "fix",
						Parser:             "closingPattern",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gitsrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				params := r.URL.Query()
				if r.URL.Path == "/api/v4/projects/123/repository/compare" {
					w.WriteHeader(200)
					assertParam(t, "from", "from", params)
					assertParam(t, "to", "to", params)
					w.Write([]byte("{\"commits\": [{\"id\": \"commit0\", \"message\": \"No reference\"}]}"))
				} else if r.URL.Path == "/api/v4/projects/123/merge_requests" {
					w.WriteHeader(200)
					assertParam(t, "state", "merged", params)
					assertParam(t, "target_branch", "main", params)
					resp := generalMRResponse(tt.mrDescription)
					w.Write([]byte(resp))
				} else {
					http.Error(w, "Not found", http.StatusNotFound)
				}
			}))
			defer gitsrv.Close()

			gc, err := clients.NewGitLab(gitsrv.URL, "token")
			assert.NoError(t, err, "NewGit error must be nil")

			gp, err := repo.New(gc, tt.issuePatterns,
				repo.WithCustomPattern(`\[(?P<scope>[^\]]*)\](?P<subject>.*)`),
				repo.WithKeepCCWithoutScope(tt.keepCCWithoutScope))
			assert.NoError(t, err, "NewGit error must be nil")
			gotCds, err := gp.GetParsedRecords("123", "from", "to", "main")

			assert.NoError(t, err, "Parse error must be nil")
			assert.Equal(t, len(tt.wantParsedRepoRecords), len(gotCds))

			for _, want := range tt.wantParsedRepoRecords {
				idx := want.idx
				want.cd.RepoRecord = generalMRRepoRecord(tt.mrDescription)
				assert.Equal(t, want.cd, gotCds[idx])
			}
		})
	}
}

func ptrTimeDate(t time.Time) *time.Time {
	return &t
}

func generalCommitRepoRecord(id int, msg string) entities.RepoRecord {
	title := strings.Split(msg, "\n")[0]
	return entities.RepoRecord{
		ID:        fmt.Sprintf("commit%d", id),
		Message:   msg,
		ShortID:   fmt.Sprintf("short_id%d", id),
		Title:     title,
		CreatedAt: ptrTimeDate(time.Date(2021, 1, 1+id, 0, 0, 0, 0, time.UTC)),
		Origin:    "commit",
		WebURL:    fmt.Sprintf("http://gitlab.example.com/my/repo/commit/%d", id),
	}
}

func generalMRRepoRecord(msg string) entities.RepoRecord {
	return entities.RepoRecord{
		ID:        "10",
		Message:   msg,
		ShortID:   "1",
		Title:     msg,
		CreatedAt: ptrTimeDate(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)),
		Origin:    "merge_request",
		WebURL:    "http://gitlab.example.com/my/repo/merge_request/1",
	}
}

func generalMRResponse(desc string) string {
	return "[\n" +
		"  {\n" +
		"    \"id\": 10,\n" +
		"    \"iid\": 1,\n" +
		"    \"title\": \"" + desc + "\",\n" +
		"    \"sha\": \"commit0\",\n" +
		"    \"description\": \"" + desc + "\",\n" +
		"    \"web_url\": \"http://gitlab.example.com/my/repo/merge_request/1\",\n" +
		"    \"merged_at\": \"2021-01-01T00:00:00Z\"\n" +
		"  }\n" +
		"]"
}

func assertParam(t *testing.T, key, value string, params url.Values) {
	p, ok := params[key]
	if !ok || ok && p[0] != value {
		t.Errorf("%s parameter not found or not as expected: want %s got %s", key, value, p[0])
	}
}

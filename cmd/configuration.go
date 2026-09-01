package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/happyagosmith/jig/internal/entities"
	"github.com/happyagosmith/jig/internal/issuetrackers"
	"github.com/happyagosmith/jig/internal/parsers"
	"github.com/happyagosmith/jig/internal/repo"
	"github.com/happyagosmith/jig/internal/repo/clients"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

type issuePatternsValue []parsers.IssuePattern

func (v *issuePatternsValue) Set(value string) error {
	var issuePattern parsers.IssuePattern
	err := yaml.Unmarshal([]byte(value), &issuePattern)
	if err != nil {
		return err
	}
	*v = append(*v, issuePattern)
	return nil
}

func (v *issuePatternsValue) Type() string {
	return "string"
}

func (v *issuePatternsValue) String() string {
	b, _ := yaml.Marshal(v)
	return string(b)
}

const (
	CustomCommitPattern     = "customCommitPattern"
	GitProvider             = "gitProvider"
	GitLabURL               = "CI_GITLAB_URL"
	GitLabToken             = "CI_GITLAB_TOKEN"
	GitHubToken             = "githubToken"
	GitHubURL               = "CI_GITHUB_URL"
	GitHubAppID             = "CI_GITHUB_APP_ID"
	GitHubInstallationID    = "CI_GITHUB_INSTALLATION_ID"
	GitHubSecret            = "CI_GITHUB_SECRET"
	GitMRBranch             = "gitMRBranch"
	JiraURL                 = "CI_JIRA_URL"
	JiraUsername            = "CI_JIRA_USERNAME"
	JiraPassword            = "CI_JIRA_PASSWORD"
	JiraClosedFeatureFilter = "jiraClosedFeatureFilter"
	JiraFixedBugFilter      = "jiraFixedBugFilter"
	JiraKnownIssuesJQL      = "jiraKnownIssuesJQL"
	IssuePatterns           = "issuePatterns"
	WithCCWithoutScope      = "withCCWithoutScope"
)

func GetConfigString(key string) string {
	return viper.GetString(key)
}

func GetConfigBool(key string) bool {
	return viper.GetBool(key)
}

func GetConfigInt64(key string) int64 {
	return viper.GetInt64(key)
}

func GetIssuePatterns() []parsers.IssuePattern {
	var issuePatterns []parsers.IssuePattern

	// First try to unmarshal as a structured array (config file format)
	err := viper.UnmarshalKey(IssuePatterns, &issuePatterns)
	if err == nil && len(issuePatterns) > 0 {
		return issuePatterns
	}

	// Fallback: try to parse as string (flag format - for backwards compatibility)
	str := viper.GetString(IssuePatterns)
	if str != "" {
		var patternsFromString issuePatternsValue
		err = yaml.Unmarshal([]byte(str), &patternsFromString)
		if err != nil {
			fmt.Printf("error unmarshaling issue patterns from string: %v\n", err)
			return nil
		}
		return patternsFromString
	}

	fmt.Printf("error unmarshaling issue patterns: %v\n", err)
	return nil
}

var cfgFile string

func InitConfiguration(cmd *cobra.Command) {
	cobra.OnInitialize(initConfig)
	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.jig.yaml)")
	cmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	issuePatterns := issuePatternsValue{
		{
			IssueTracker: "silk",
			Pattern:      `SILK-\d+|silk-\d+`,
		},
		{
			IssueTracker: "jira",
			Pattern:      `[A-Z]+-\d+`,
		},
		{
			IssueTracker: "jira",
			Pattern:      `j_(.+)`,
		},
		{
			IssueTracker: "git",
			Pattern:      `#(\d+)`,
		},
	}
	cmd.PersistentFlags().Var(&issuePatterns, IssuePatterns, "Issue patterns used to determine the issue tracker associated with each issue key")
	viper.BindPFlag(IssuePatterns, cmd.PersistentFlags().Lookup(IssuePatterns))

	cmd.PersistentFlags().String(GitMRBranch, "", "The branch for which the merge request is being parsed. If this is not specified, the merge requests will not be processed.")
	viper.BindPFlag(GitMRBranch, cmd.PersistentFlags().Lookup(GitMRBranch))

	cmd.PersistentFlags().Bool(WithCCWithoutScope, false, "if true, extract conventional commit without scope")
	viper.BindPFlag(WithCCWithoutScope, cmd.PersistentFlags().Lookup(WithCCWithoutScope))

	cmd.PersistentFlags().String(CustomCommitPattern, `\[(?P<scope>[^\]]*)\](?P<subject>.*)`, "Custom pattern to apply on the commit and merge request title to extract the issue keys and the summary. If the message is not a conventional commit message, this custom pattern is applied. The pattern should include the named groups scope and subject")
	viper.BindPFlag(CustomCommitPattern, cmd.PersistentFlags().Lookup(CustomCommitPattern))

	cmd.PersistentFlags().String(GitLabURL, "", "GitLab base URL")
	viper.BindPFlag(GitLabURL, cmd.PersistentFlags().Lookup(GitLabURL))

	cmd.PersistentFlags().String(GitLabToken, "", "GitLab token with read REST API permissions")
	viper.BindPFlag(GitLabToken, cmd.PersistentFlags().Lookup(GitLabToken))

	cmd.PersistentFlags().String(GitHubToken, "", "GitHub token with read REST API permissions")
	viper.BindPFlag(GitHubToken, cmd.PersistentFlags().Lookup(GitHubToken))

	cmd.PersistentFlags().String(GitHubURL, "", "GitHub base URL (for GitHub Enterprise, e.g. https://github.example.com)")
	viper.BindPFlag(GitHubURL, cmd.PersistentFlags().Lookup(GitHubURL))

	cmd.PersistentFlags().Int64(GitHubAppID, 0, "GitHub App ID (for GitHub App authentication)")
	viper.BindPFlag(GitHubAppID, cmd.PersistentFlags().Lookup(GitHubAppID))

	cmd.PersistentFlags().Int64(GitHubInstallationID, 0, "GitHub App installation ID")
	viper.BindPFlag(GitHubInstallationID, cmd.PersistentFlags().Lookup(GitHubInstallationID))

	cmd.PersistentFlags().String(GitHubSecret, "", "GitHub App private key PEM content")
	viper.BindPFlag(GitHubSecret, cmd.PersistentFlags().Lookup(GitHubSecret))

	cmd.PersistentFlags().String(JiraURL, "", "Jira base URL")
	viper.BindPFlag(JiraURL, cmd.PersistentFlags().Lookup(JiraURL))

	cmd.PersistentFlags().String(JiraUsername, "", "Jira username with read REST API permissions")
	viper.BindPFlag(JiraUsername, cmd.PersistentFlags().Lookup(JiraUsername))

	cmd.PersistentFlags().String(JiraPassword, "", "Jira password/token with read REST API permissions")
	viper.BindPFlag(JiraPassword, cmd.PersistentFlags().Lookup(JiraPassword))

	cmd.PersistentFlags().String(JiraClosedFeatureFilter, "Story:GOLIVE,TECH TASK:Completata", "List of filters type:status that identify the closed features")
	viper.BindPFlag(JiraClosedFeatureFilter, cmd.PersistentFlags().Lookup(JiraClosedFeatureFilter))

	cmd.PersistentFlags().String(JiraFixedBugFilter, "BUG:FIXED,BUG:RELEASED", "List of filters type:status that identify the fixed bugs")
	viper.BindPFlag(JiraFixedBugFilter, cmd.PersistentFlags().Lookup(JiraFixedBugFilter))

	cmd.PersistentFlags().String(JiraKnownIssuesJQL, "status not in (Done, RELEASED, Fixed, GOLIVE, Cancelled) AND issuetype in (Bug, \"TECH DEBT\")", "Jira JQL to retrieve the known issues")
	viper.BindPFlag(JiraKnownIssuesJQL, cmd.PersistentFlags().Lookup(JiraKnownIssuesJQL))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".jig")
	}

	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "using config file:", viper.ConfigFileUsed())
	}
}

func addJiraOpt(label string, value string, opts *[]issuetrackers.JiraOpt, opt func(string, string) issuetrackers.JiraOpt) error {
	fmt.Printf("using %s -> %s\n", label, value)
	filters := strings.Split(value, ",")
	if len(filters) == 0 {
		return fmt.Errorf("wrong format of %s, expected list type:status separated by coma", label)
	}
	for _, cff := range filters {
		f := strings.Split(cff, ":")
		if len(f) != 2 {
			return fmt.Errorf("wrong format of %s, expected list type:status separated by coma", label)
		}
		*opts = append(*opts, opt(strings.TrimSpace(f[0]), strings.TrimSpace(f[1])))
	}

	return nil
}

func ConfigureJira() (*issuetrackers.Jira, error) {
	if GetConfigString(JiraURL) == "" || GetConfigString(JiraUsername) == "" || GetConfigString(JiraPassword) == "" {
		return nil, fmt.Errorf("%s, %s and %s are required", JiraURL, JiraUsername, JiraPassword)
	}

	var opts []issuetrackers.JiraOpt
	addJiraOpt("jiraClosedFeatureFilter", GetConfigString(JiraClosedFeatureFilter), &opts, issuetrackers.WithClosedFeatureFilter)
	addJiraOpt("jiraFixedBugFilter", GetConfigString(JiraFixedBugFilter), &opts, issuetrackers.WithFixedBugFilter)
	fmt.Printf("using %s -> %s\n", "jiraKnownIssuesJQL", GetConfigString(JiraKnownIssuesJQL))
	fmt.Printf("using %s -> %s\n", JiraURL, GetConfigString(JiraURL))

	opts = append(opts, issuetrackers.WithKnownIssueJql(GetConfigString(JiraKnownIssuesJQL)))
	jiraTracker, err := issuetrackers.NewJira(
		GetConfigString(JiraURL),
		GetConfigString(JiraUsername),
		GetConfigString(JiraPassword),
		opts...,
	)

	return &jiraTracker, err
}

// ConfigureRepoTrackers builds one client per configured provider and returns
// them in a map keyed by provider name ("gitlab", "github"). The default
// provider (used for repos without an explicit gitProvider field) is also
// returned so callers can fall back to it.
func ConfigureRepoTrackers() (trackers map[string]entities.RepoTracker, defaultProvider string, err error) {
	trackers = make(map[string]entities.RepoTracker)

	// GitLab — configured when CI_GITLAB_URL + CI_GITLAB_TOKEN are present
	gitlabURL := GetConfigString(GitLabURL)
	gitlabToken := GetConfigString(GitLabToken)
	if gitlabURL != "" && gitlabToken != "" {
		fmt.Printf("using %s -> %s\n", GitLabURL, gitlabURL)
		gl, err := clients.NewGitLab(gitlabURL, gitlabToken)
		if err != nil {
			return nil, "", err
		}
		trackers["gitlab"] = gl
	}

	appID := GetConfigInt64(GitHubAppID)
	installationID := GetConfigInt64(GitHubInstallationID)
	if appID != 0 && installationID != 0 {
		secret := GetConfigString(GitHubSecret)
		if secret == "" {
			return nil, "", fmt.Errorf("%s is required for GitHub App authentication", GitHubSecret)
		}
		privateKey := []byte(secret)
		fmt.Printf("using %s -> github app\n", GitHubAppID)
		gh, err := clients.NewGitHubApp(appID, installationID, privateKey, GetConfigString(GitHubURL))
		if err != nil {
			return nil, "", err
		}
		trackers["github"] = gh
	} else {
		// Fallback to token authentication
		ghToken := GetConfigString(GitHubToken)
		if ghToken == "" {
			ghToken = os.Getenv("GITHUB_TOKEN")
		}
		if ghToken != "" {
			fmt.Printf("using githubToken -> github\n")
			gh, err := clients.NewGitHub(ghToken)
			if err != nil {
				return nil, "", err
			}
			trackers["github"] = gh
		}
	}

	if len(trackers) == 0 {
		return nil, "", fmt.Errorf("no git provider configured: set %s+%s for GitLab or %s for GitHub", GitLabURL, GitLabToken, GitHubToken)
	}

	// Determine default: explicit gitProvider config wins, otherwise pick the
	// only configured one (or "gitlab" when both are present).
	switch GetConfigString(GitProvider) {
	case "github":
		if _, ok := trackers["github"]; !ok {
			return nil, "", fmt.Errorf("gitProvider is \"github\" but githubToken is not set")
		}
		defaultProvider = "github"
	default:
		if _, ok := trackers["gitlab"]; ok {
			defaultProvider = "gitlab"
		} else {
			defaultProvider = "github"
		}
	}

	return trackers, defaultProvider, nil
}

func ConfigureRepoServices(trackers map[string]entities.RepoTracker) (map[string]entities.RepoService, error) {
	fmt.Printf("using %s -> %s\n", CustomCommitPattern, GetConfigString(CustomCommitPattern))
	fmt.Printf("using %s -> %v\n", WithCCWithoutScope, GetConfigString(WithCCWithoutScope))
	fmt.Printf("using %s -> %v\n", GitMRBranch, GetConfigString(GitMRBranch))

	services := make(map[string]entities.RepoService, len(trackers))
	for provider, client := range trackers {
		svc, err := repo.New(client, GetIssuePatterns(),
			repo.WithDefaultMRBranch(GetConfigString(GitMRBranch)),
			repo.WithCustomPattern(GetConfigString(CustomCommitPattern)),
			repo.WithKeepCCWithoutScope(GetConfigBool(WithCCWithoutScope)))
		if err != nil {
			return nil, err
		}
		services[provider] = svc
	}
	return services, nil
}

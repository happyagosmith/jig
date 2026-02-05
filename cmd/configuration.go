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
	GitURL                  = "gitURL"
	GitToken                = "gitToken"
	GitMRBranch             = "gitMRBranch"
	JiraURL                 = "jiraURL"
	JiraUsername            = "jiraUsername"
	JiraPassword            = "jiraPassword"
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

	cmd.PersistentFlags().String(GitURL, "", "Git base URL")
	viper.BindPFlag(GitURL, cmd.PersistentFlags().Lookup(GitURL))

	cmd.PersistentFlags().String(GitToken, "", "Git token with read REST API permissions")
	viper.BindPFlag(GitToken, cmd.PersistentFlags().Lookup(GitToken))

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
		*opts = append(*opts, opt(f[0], f[1]))
	}

	return nil
}

func ConfigureJira() (*issuetrackers.Jira, error) {
	if GetConfigString(JiraURL) == "" || GetConfigString(JiraUsername) == "" || GetConfigString(JiraPassword) == "" {
		return nil, fmt.Errorf("jiraURL, jiraUsername and jiraPassword are required")
	}

	var opts []issuetrackers.JiraOpt
	addJiraOpt("jiraClosedFeatureFilter", GetConfigString(JiraClosedFeatureFilter), &opts, issuetrackers.WithClosedFeatureFilter)
	addJiraOpt("jiraFixedBugFilter", GetConfigString(JiraFixedBugFilter), &opts, issuetrackers.WithFixedBugFilter)
	fmt.Printf("using %s -> %s\n", "jiraKnownIssuesJQL", GetConfigString(JiraKnownIssuesJQL))
	fmt.Printf("using %s -> %s\n", "jiraURL", GetConfigString(JiraURL))

	opts = append(opts, issuetrackers.WithKnownIssueJql(GetConfigString(JiraKnownIssuesJQL)))
	jiraTracker, err := issuetrackers.NewJira(
		GetConfigString(JiraURL),
		GetConfigString(JiraUsername),
		GetConfigString(JiraPassword),
		opts...,
	)

	return &jiraTracker, err
}

func ConfigureRepoTracker() (entities.RepoTracker, error) {
	if GetConfigString(GitURL) == "" || GetConfigString(GitToken) == "" {
		return nil, fmt.Errorf("gitURL and gitToken are required")
	}
	fmt.Printf("using %s -> %s\n", "gitURL", GetConfigString(GitURL))

	git, err := clients.NewGitLab(GetConfigString(GitURL), GetConfigString(GitToken))

	return git, err
}

func ConfigureRepoService(repoClient entities.RepoClient) (entities.RepoService, error) {
	fmt.Printf("using %s -> %s\n", CustomCommitPattern, GetConfigString(CustomCommitPattern))
	fmt.Printf("using %s -> %v\n", WithCCWithoutScope, GetConfigString(WithCCWithoutScope))
	fmt.Printf("using %s -> %v\n", GitMRBranch, GetConfigString(GitMRBranch))

	repoParser, err := repo.New(repoClient, GetIssuePatterns(),
		repo.WithDefaultMRBranch(GetConfigString(GitMRBranch)),
		repo.WithCustomPattern(GetConfigString(CustomCommitPattern)),
		repo.WithKeepCCWithoutScope(GetConfigBool(WithCCWithoutScope)))

	return repoParser, err
}

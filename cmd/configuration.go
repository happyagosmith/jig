package cmd

import (
	"fmt"
	"strings"

	"github.com/happyagosmith/jig/internal/model"
	"github.com/happyagosmith/jig/internal/parsers"
	git "github.com/happyagosmith/jig/internal/repositories"
	"github.com/happyagosmith/jig/internal/trackers"
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
	var issuePatterns issuePatternsValue

	str := viper.GetString(IssuePatterns)
	err := yaml.Unmarshal([]byte(str), &issuePatterns)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return nil
	}

	return issuePatterns
}

var initialized bool
var cfgFile string

func InitConfiguration() {
	if initialized {
		return
	}

	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.jig.yaml)")
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	issuePatterns := issuePatternsValue{
		{
			IssueTracker: "silk",
			Pattern:      `SILK-\d+|silk-\d+`,
		},
		{
			IssueTracker: "jira",
			Pattern:      `j_(.+)|[A-Z]+-\d+`,
		},
		{
			IssueTracker: "git",
			Pattern:      `#(\d+)`,
		},
	}
	rootCmd.PersistentFlags().Var(&issuePatterns, IssuePatterns, "Issue patterns used to determine the issue tracker associated with each issue key")
	viper.BindPFlag(IssuePatterns, rootCmd.PersistentFlags().Lookup(IssuePatterns))

	rootCmd.PersistentFlags().Bool(WithCCWithoutScope, false, "if true, extract conventional commit without scope")
	viper.BindPFlag(WithCCWithoutScope, rootCmd.PersistentFlags().Lookup(WithCCWithoutScope))

	rootCmd.PersistentFlags().String(CustomCommitPattern, `\[(?P<scope>[^\]]*)\](?P<subject>.*)`, "Custom pattern to apply on the git commit message to extract the issue keys and the summary. If the message is not a conventional commit message, this custom pattern is applied. The pattern should include the named groups scope and subject")
	viper.BindPFlag(CustomCommitPattern, rootCmd.PersistentFlags().Lookup(CustomCommitPattern))

	rootCmd.PersistentFlags().String(GitURL, "", "Git base URL")
	viper.BindPFlag(GitURL, rootCmd.PersistentFlags().Lookup(GitURL))

	rootCmd.PersistentFlags().String(GitToken, "", "Git token with read REST API permissions")
	viper.BindPFlag(GitToken, rootCmd.PersistentFlags().Lookup(GitToken))

	rootCmd.PersistentFlags().String(JiraURL, "", "Jira base URL")
	viper.BindPFlag(JiraURL, rootCmd.PersistentFlags().Lookup(JiraURL))

	rootCmd.PersistentFlags().String(JiraUsername, "", "Jira username with read REST API permissions")
	viper.BindPFlag(JiraUsername, rootCmd.PersistentFlags().Lookup(JiraUsername))

	rootCmd.PersistentFlags().String(JiraPassword, "", "Jira password/token with read REST API permissions")
	viper.BindPFlag(JiraPassword, rootCmd.PersistentFlags().Lookup(JiraPassword))

	rootCmd.PersistentFlags().String(JiraClosedFeatureFilter, "Story:GOLIVE,TECH TASK:Completata", "List of filters type:status that identify the closed features")
	viper.BindPFlag(JiraClosedFeatureFilter, rootCmd.PersistentFlags().Lookup(JiraClosedFeatureFilter))

	rootCmd.PersistentFlags().String(JiraFixedBugFilter, "BUG:FIXED,BUG:RELEASED", "List of filters type:status that identify the fixed bugs")
	viper.BindPFlag(JiraFixedBugFilter, rootCmd.PersistentFlags().Lookup(JiraFixedBugFilter))

	rootCmd.PersistentFlags().String(JiraKnownIssuesJQL, "status not in (Done, RELEASED, Fixed, GOLIVE, Cancelled) AND issuetype in (Bug, \"TECH DEBT\")", "Jira JQL to retrieve the known issues")
	viper.BindPFlag(JiraKnownIssuesJQL, rootCmd.PersistentFlags().Lookup(JiraKnownIssuesJQL))
}

func addJiraOpt(label string, value string, opts *[]trackers.JiraOpt, opt func(string, string) trackers.JiraOpt) {
	fmt.Printf("using %s -> %s\n", label, value)
	filters := strings.Split(value, ",")
	if len(filters) == 0 {
		CheckErr(fmt.Errorf("wrong format of %s, expected list type:status separated by coma", label))
	}
	for _, cff := range filters {
		f := strings.Split(cff, ":")
		if len(f) != 2 {
			CheckErr(fmt.Errorf("wrong format of %s, expected list type:status separated by coma", label))
		}
		*opts = append(*opts, opt(f[0], f[1]))
	}
}

func ConfigureJira() trackers.Jira {
	if GetConfigString(JiraURL) == "" || GetConfigString(JiraUsername) == "" || GetConfigString(JiraPassword) == "" {
		CheckErr(fmt.Errorf("jiraURL, jiraUsername and jiraPassword are required"))
	}

	var opts []trackers.JiraOpt
	addJiraOpt("jiraClosedFeatureFilter", GetConfigString(JiraClosedFeatureFilter), &opts, trackers.WithClosedFeatureFilter)
	addJiraOpt("jiraFixedBugFilter", GetConfigString(JiraFixedBugFilter), &opts, trackers.WithFixedBugFilter)
	fmt.Printf("using %s -> %s\n", "jiraKnownIssuesJQL", GetConfigString(JiraKnownIssuesJQL))
	fmt.Printf("using %s -> %s\n", "jiraURL", GetConfigString(JiraURL))

	opts = append(opts, trackers.WithKnownIssueJql(GetConfigString(JiraKnownIssuesJQL)))
	jiraTracker, err := trackers.NewJira(
		GetConfigString(JiraURL),
		GetConfigString(JiraUsername),
		GetConfigString(JiraPassword),
		opts...,
	)
	CheckErr(err)

	return jiraTracker
}

func ConfigureRepoSRV() (model.RepoSRV, error) {
	if GetConfigString(GitURL) == "" || GetConfigString(GitToken) == "" {
		CheckErr(fmt.Errorf("gitURL and gitToken are required"))
	}
	fmt.Printf("using %s -> %s\n", "gitURL", GetConfigString(GitURL))
	fmt.Printf("using %s -> %s\n", "customCommitPattern", GetConfigString(CustomCommitPattern))
	fmt.Printf("using %s -> %v\n", "withCCWithoutScope", GetConfigString(WithCCWithoutScope))

	git, err := git.NewClient(GetConfigString(GitURL), GetConfigString(GitToken), GetIssuePatterns(),
		git.WithCustomPattern(GetConfigString(CustomCommitPattern)), git.WithKeepCCWithoutScope(GetConfigBool(WithCCWithoutScope)))
	if err != nil {
		return nil, err
	}
	CheckErr(err)

	return git, nil
}

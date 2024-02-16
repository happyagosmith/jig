/*
Copyright © 2023 Happy Smith happyagosmith@gmail.com
*/
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/happyagosmith/jig/internal/model"
	"github.com/happyagosmith/jig/internal/trackers"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func addJiraOpt(label string, value string, opts *[]trackers.JiraOpt, opt func(string, string) trackers.JiraOpt) {
	fmt.Printf("Using %s -> %s\n", label, value)
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

func EnrichModel(b []byte) []byte {
	var opts []trackers.JiraOpt
	fmt.Printf("Using %s -> %s\n", "issuePattern in GIT commit messages", viper.GetString("issuePattern"))
	addJiraOpt("jiraClosedFeatureFilter", viper.GetString("jiraClosedFeatureFilter"), &opts, trackers.WithClosedFeatureFilter)
	addJiraOpt("jiraFixedBugFilter", viper.GetString("jiraFixedBugFilter"), &opts, trackers.WithFixedBugFilter)

	opts = append(opts, trackers.WithKnownIssueJql(viper.GetString("jiraKnownIssuesJQL")))
	jiraTracker, err := trackers.NewJira(
    viper.GetString("jiraURL"),
    viper.GetString("jiraUsername"),
    viper.GetString("jiraPassword"),
    opts...,
	)
	CheckErr(err)

	model, err := model.New(b, model.WithIssueTracker(jiraTracker))
	CheckErr(err)

	err = model.EnrichWithGit(
    viper.GetString("gitURL"),
    viper.GetString("gitToken"),
    viper.GetString("issuePattern"),
    viper.GetString("customCommitPattern"))
	CheckErr(err)

	err = model.EnrichWithIssueTrackers()
	CheckErr(err)

	b, err = model.Yaml()
	CheckErr(err)

	return b
}

func setEnrichFlags(cmd *cobra.Command) {
	cmd.Flags().String("issuePattern", `(^j_(?P<jira_1>.*)$)|(?P<jira_2>^[^\_]+$)`, "Pattern to apply on the git commit message to extract the issue keys from the message. The pattern should include the named groups composed by noun with number (e.g. jira_1). The noun refers to the issue tracker (at the moment only jira is supported). The number has the purpose to define more than one pattern for the same issue tracker (this is usefull if the commit message format is changed over the time). The pattern must be a valid regex pattern.")
	viper.BindPFlag("issuePattern", cmd.Flags().Lookup("issuePattern"))

	cmd.Flags().String("customCommitPattern", `\[(?P<scope>[^\]]*)\](?P<subject>.*)`, "Custom pattern to apply on the git commit message to extract the issue keys and the summary. If the message is not a conventional commit message, this custom pattern is applied. The pattern should include the named groups scope and subject")
	viper.BindPFlag("customCommitPattern", cmd.Flags().Lookup("customCommitPattern"))

	cmd.Flags().String("gitURL", "", "Git base URL")
	viper.BindPFlag("gitURL", cmd.Flags().Lookup("gitURL"))

	cmd.Flags().String("gitToken", "", "Git token with read REST API permissions")
	viper.BindPFlag("gitToken", cmd.Flags().Lookup("gitToken"))

	cmd.Flags().String("jiraURL", "", "Jira base URL")
	viper.BindPFlag("jiraURL", cmd.Flags().Lookup("jiraURL"))

	cmd.Flags().String("jiraUsername", "", "Jira username with read REST API permissions")
	viper.BindPFlag("jiraUsername", cmd.Flags().Lookup("jiraUsername"))

	cmd.Flags().String("jiraPassword", "", "Jira password/token with read REST API permissions")
	viper.BindPFlag("jiraPassword", cmd.Flags().Lookup("jiraPassword"))

	cmd.Flags().String("jiraClosedFeatureFilter", "Story:GOLIVE,TECH TASK:Completata", "List of filters type:status that identify the closed features")
	viper.BindPFlag("jiraClosedFeatureFilter", cmd.Flags().Lookup("jiraClosedFeatureFilter"))

	cmd.Flags().String("jiraFixedBugFilter", "BUG:FIXED,BUG:RELEASED", "List of filters type:status that identify the fixed bugs")
	viper.BindPFlag("jiraFixedBugFilter", cmd.Flags().Lookup("jiraFixedBugFilter"))

	cmd.Flags().String("jiraKnownIssuesJQL", "status not in (Done, RELEASED, Fixed, GOLIVE, Cancelled) AND issuetype in (Bug, \"TECH DEBT\")", "Jira JQL to retrieve the known issues")
	viper.BindPFlag("jiraKnownIssuesJQL", cmd.Flags().Lookup("jiraKnownIssuesJQL"))
}

func newEnrichCmd() *cobra.Command {
	enrichCmd := &cobra.Command{
    Use:   "enrich [model.yaml]",
    Short: "Enrich the model.yaml file with the generated values extracted from Git and Jira.",
    Long: `Enrich the model.yaml file with the generated values extracted from Git and Jira.

The model.yaml file should include the list "gitRepo" with an element for each repo 
to be processed. Following an example:	
    services:
    - gitRepoID: xxx           # id of the gitLab repo
      previousVersion: 1.0.0   # tag from which process the commits
      version: 2.0.0           # tag to which process the commits 
      label: service1          # label to assign to the gitLab repo
	  
The file model.yaml will be enriched with the key "generatedValues" including the values 
extracted from Git and Jira. Following an example:

generatedValues:
    features: 
    	service1:
    	- issueKey: AAA-000
          issueSummary: 'Fix Comment from the issue tracker'
          issueType: TECH TASK
          issueStatus: Completata
          issueTrackerType: JIRA
          category: CLOSED_FEATURE
          isbreakingchange: true
    bugs:
    	service1:
    	- issueKey: AAA-111
          issueSummary: 'Fix Comment from the issue tracker'
          issueType: Bug
          issueStatus: RELEASED
          issueTrackerType: JIRA
          category: FIXED_BUG
          isBreakingChange: false
    	service2:
    	- issueKey: AAA-222
          issueSummary: 'Fix Comment from the issue tracker'
          issueType: Bug
          issueStatus: FIXED
          issueTrackerType: JIRA
          category: FIXED_BUG
          isBreakingChange: false
    knownIssues:
    	service2:
    	- issueKey: AAA-333
          issueType: TECH DEBT
          issueSummary: To be implemented
          issueStatus: Da completare
          issueTrackerType: JIRA
          category: OTHER
          isBreakingChange: false
    breakingChange: 
    	service1:
    	- issueKey: AAA-000
          issueSummary: 'Fix Comment from the issue tracker'
          issueType: TECH TASK
          issueStatus: Completata
          issueTrackerType: JIRA
          category: CLOSED_FEATURE
          isbreakingchange: true
    gitRepos:
      - gitRepoID: 1234
    	label: service1
    	previousVersion: 0.0.1
    	version: 0.0.2
    	extractedKeys:
    	- category: BUG_FIX
    	  issueKey: AAA-000
    	  summary: 'fix comment from git'
    	  message: 'fix(j_AAA-000)!: fix comment from git'
    	  issueTrackerType: JIRA
    	  isbreakingchange: true
    	- issueKey: AAA-111
    	  summary: 'fix comment from git'
    	  message: '[AAA-111] fix comment from git'
    	  issueTrackerType: JIRA
    	  isbreakingchange: false
      - gitRepoID: 5678
    	label: service2
    	jiraComponent: jComponent # used to retrieve the known issues
    	jiraProject: jProject # used to retrieve the known issues
    	previousVersion: 1.2.0
    	version: 1.2.1
    	extractedKeys:
    	- category: BUG_FIX
    	  issueKey: AAA-222
    	  summary: 'fix comment from git'
    	  message: 'fix(j_AAA-222): fix comment from git'
    	  issueTrackerType: JIRA
    	  isbreakingchange: false
    	- issueKey: AAA-333
    	  summary: 'comment from git'
    	  message: '[AAA-333] comment from git'
    	  issueTrackerType: JIRA
    	  isbreakingchange: false
`,
    Args: func(cmd *cobra.Command, args []string) error {
    	return cobra.ExactArgs(1)(cmd, args)
    },
    RunE: func(cmd *cobra.Command, args []string) error {
    	v, err := os.ReadFile(args[0])
    	CheckErr(err)

    	b := EnrichModel(v)
    	err = os.WriteFile(args[0], b, 0644)
    	CheckErr(err)

    	return nil
    },
	}
	setEnrichFlags(enrichCmd)
	return enrichCmd
}

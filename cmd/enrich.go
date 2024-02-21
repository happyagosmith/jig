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

func ConfigureJira() trackers.Jira {
	jiraURL := viper.GetString("jiraURL")
	jiraUsername := viper.GetString("jiraUsername")
	jiraPassword := viper.GetString("jiraPassword")

	if jiraURL == "" || jiraUsername == "" || jiraPassword == "" {
		CheckErr(fmt.Errorf("jiraURL, jiraUsername and jiraPassword are required"))
	}

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

	return jiraTracker
}

func EnrichModel(b []byte) []byte {
	jiraTracker := ConfigureJira()
	model, err := model.New(b, model.WithIssueTracker(jiraTracker))
	CheckErr(err)

	gitURL := viper.GetString("gitURL")
	gitToken := viper.GetString("gitToken")

	if gitURL == "" || gitToken == "" {
		CheckErr(fmt.Errorf("gitURL and gitToken are required"))
	}
	fmt.Printf("Using %s -> %s\n", "gitURL", gitURL)
	fmt.Printf("Using %s -> %s\n", "gitToken", gitToken)

	err = model.EnrichWithGit(
		gitURL,
		gitToken,
		viper.GetString("issuePattern"),
		viper.GetString("customCommitPattern"))
	CheckErr(err)

	err = model.EnrichWithIssueTrackers()
	CheckErr(err)

	b, err = model.Yaml()
	CheckErr(err)

	return b
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

	return enrichCmd
}

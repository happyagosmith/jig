/*
Copyright © 2023 Happy Smith happyagosmith@gmail.com
*/
package cmd

import (
	"os"

	"github.com/happyagosmith/jig/internal/filehandler/model"
	"github.com/spf13/cobra"
)

func EnrichModel(cmd *cobra.Command, b []byte) []byte {
	jiratracker, err := ConfigureJira()
	CheckErr(cmd, err)
	repotracker, err := ConfigureRepoClient()
	CheckErr(cmd, err)
	repoparser, err := ConfigureRepoParser()
	CheckErr(cmd, err)

	model, err := model.New(b,
		model.WithRepoClient(repotracker),
		model.WithRepoParser(repoparser),
		model.WithIssueTracker("JIRA", jiratracker),
		model.WithIssueTracker("GIT", repotracker),
		model.WithIssueTracker("SILK", nil),
	)
	CheckErr(cmd, err)

	err = model.EnrichWithRepos()
	CheckErr(cmd, err)

	err = model.EnrichWithIssueTrackers()
	CheckErr(cmd, err)

	b, err = model.Yaml()
	CheckErr(cmd, err)

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
			modelPath := args[0]
			fl := NewFileLoader(GetConfigString(GitToken))

			cmd.Printf("using model file: %s\n", modelPath)
			v, err := fl.GetFile(modelPath)
			CheckErr(cmd, err)

			b := EnrichModel(cmd, v)
			err = os.WriteFile(modelPath, b, 0644)
			CheckErr(cmd, err)

			return nil
		},
	}

	return enrichCmd
}

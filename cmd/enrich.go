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
	cmd.Flags().String("issuePattern", `(^j_(?P<jira_1>.*)$)|(?P<jira_2>^[^\_]+$)`, "Pattern to apply on the git commit message to extract the issue keys from the message. The pattern should include the named groups composed by noun with number (e.g. jira1). The noun refers to the issue tracker (at the moment only jira is supproted). The number is allowed in orde to define more than one pattern for the same issu tracker (this is usefull if the commit message format is changed over the time). The pattern should be a valid regex pattern.")
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
		Short: "enrich [model.yaml]",
		Long: `Enrich the model.yaml file with the generated values extracted from Git and Jira.

The model.yaml file should include the list "gitRepo" with an element for each repo 
to be processed. Following an example:	
    gitRepos:
    - ID: xxx          # id of the gitLab repo
      fromTag: 1.0.0   # tag from which process the commits
      toTag: 2.0.0     # tag to which process the commits 
      label: repoLabel # label to assign to the gitLab repo
	  
The file model.yaml will be enriched with the key "generatedValues" including the values 
extracted from Git and Jira. Following an example:

    generatedValues:
       features:
       - category: repoLabel
         parent: XXX-1234
       	 key: XXX-AAAA
         summary: This is a feature
         status: Completata
         type: TECH TASK
       bugs:
       - category: repoLabel
         parent: XXX-1234
         key: XXX-BBBB
         summary: 'This is a bug'
         status: RELEASED
         type: Bug
       knownIssues:
       - category: repoLabel
         parent: XXX-1234
         key: XXX-CCCC
         summary: "this is a known issue
         status: Creato
         type: Bug
      gitRepos:
      - label: repoLabel
        extractedKeys:
        - XXX-AAAA
        - XXX-BBBB
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

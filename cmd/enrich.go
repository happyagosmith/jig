/*
Copyright © 2023 Happy Smith happyagosmith@gmail.com
*/
package cmd

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/happyagosmith/jig/internal/filehandler/model"
	"github.com/spf13/cobra"
)

func EnrichModel(cmd *cobra.Command, b []byte) []byte {
	jiraTracker, err := ConfigureJira()
	CheckErr(cmd, err)

	repoTracker, err := ConfigureRepoTracker()
	CheckErr(cmd, err)

	repoService, err := ConfigureRepoService(repoTracker)
	CheckErr(cmd, err)

	model, err := model.New(b,
		model.WithRepoService(repoService),
		model.WithIssueTracker("JIRA", jiraTracker),
		model.WithIssueTracker("GIT", repoTracker),
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

//go:embed testdata/model.yaml
var modelExample []byte

//go:embed testdata/model-enriched.yaml
var enrichedModelExample []byte

func newEnrichCmd() *cobra.Command {

	enrichCmd := &cobra.Command{
		Use:   "enrich [model.yaml]",
		Short: "Enrich the model.yaml file with the generated values extracted from Git and Jira.",
		Long: fmt.Sprintf(`Enrich the model.yaml file with the generated values extracted from Git and Jira.

The model.yaml file should include the list "gitRepo" with an element for each repo 
to be processed. Following an example:	
%s
	  
The file model.yaml will be enriched with the key "generatedValues" including the values 
extracted from Git and Jira. Following an example:

%s
`, modelExample, enrichedModelExample),
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

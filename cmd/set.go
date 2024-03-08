/*
Copyright © 2023 Happy Smith happyagosmith@gmail.com
*/
package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/happyagosmith/jig/internal/model"
)

func setVersions() *cobra.Command {
	updateCmd := &cobra.Command{
		Use:   "setVersions",
		Short: "setVersions [model.yaml]",
		Long: `The command is designed to automate the process of updating version information 
in the [model.yaml] file. 

This command works in conjunction with the checkVersion field in your model. The checkVersion 
field should specify the file and the YAML path from which the version information should be 
read. An example configuration might look like this:

services:
	- gitRepoID: 1234
	  label: service1
	  previousVersion: 0.0.1
	  version: 0.0.2
	  checkVersion: '@filepath:$.versions.a'
	- gitRepoID: 5678
	  jiraComponent: jComponent # used to retrieve the known issues from jira
	  jiraProject: jProject # used to retrieve the known issues from jira
	  label: service2
	  previousVersion: 1.2.0
	  version: 1.2.1
	  checkVersion: '@filepath:$.a[?(@.b == ''label'')].c'

When you run jig setVersions, the command reads the version information from the specified file 
and YAML path and updates the version field in your model with this information. 

If the version field is updated with a new value, the previousVersion field is also updated. 
The previousVersion field will hold the value that was previously in the version field before 
the update.

This command is particularly useful for keeping your version information up to date without 
having to manually change the version and previousVersion fields each time a new version 
is released.

Besides, the model.yaml file will include the "gitRepoURL" and the "gitReleaseURL"  for each repo 
as well. Following an example:	
services:
	- gitRepoID: 1234
	  label: service1
	  previousVersion: 0.0.1
	  version: 0.0.2
	  checkVersion: '@filepath:$.versions.a'
	  gitRepoURL: https://repo-service1-url
      gitReleaseURL: https://repo-service1-url/-/releases/0.0.2
	- gitRepoID: 5678
	  jiraComponent: jComponent # used to retrieve the known issues from jira
	  jiraProject: jProject # used to retrieve the known issues from jira
	  label: service2
	  previousVersion: 1.2.0
	  version: 1.2.1
	  checkVersion: '@filepath:$.a[?(@.b == ''label'')].c'
	  gitRepoURL: https://repo-service1-url/-/releases/1.2.1
      gitReleaseURL: https://repo-service1-url/-/releases/1.2.1
`,
		Args: func(cmd *cobra.Command, args []string) error {
			return cobra.ExactArgs(1)(cmd, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			modelPath := args[0]

			fl := NewFileLoader(GetConfigString(GitToken))
			cmd.Printf("using model file: %s\n", modelPath)
			b, err := fl.GetFile(modelPath)
			CheckErr(err)

			vcs, err := ConfigureRepoSRV()
			CheckErr(err)

			m, err := model.New(b, model.WithRepoSRV(vcs))
			CheckErr(err)

			err = m.SetVersions(filepath.Dir(modelPath))
			CheckErr(err)

			err = m.SetReposInfos()
			CheckErr(err)

			b, err = m.Yaml()
			CheckErr(err)

			err = os.WriteFile(modelPath, b, 0644)
			CheckErr(err)

			cmd.Printf("\nversions updated with success in the model %s\n", modelPath)

			return nil
		},
	}

	return updateCmd
}

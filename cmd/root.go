/*
Copyright © 2023 Happy Smith happyagosmith@gmail.com

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func NewRootCmd(version string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "jig",
		Short: "Jig is a tool designed to automate the creation of release notes.",
		Long: `Jig is a tool designed to automate the creation of release notes for 
software products composed of one or more components, each with its own Git 
repository. Jig leverages the information found in commit messages across these 
repositories and enriches it with details from the issue tracker (e.g. Jira).

Jig has two main objectives: firstly, to augment a model.yaml file, and secondly, 
to generate a release note that is based on the improved model.yaml file and 
conforms to a particular template.
	`,
		Run: func(cmd *cobra.Command, args []string) {
			if v, _ := cmd.Flags().GetBool("version"); v {
				fmt.Println("jig version " + version)
				return
			}

			cmd.Help()
		},
	}
	InitConfiguration(rootCmd)

	rootCmd.AddCommand(newEnrichCmd())
	rootCmd.AddCommand(newGenerateCmd())
	rootCmd.AddCommand(newSetCmd())

	return rootCmd
}

func CheckErr(cmd *cobra.Command, e error) {
	if e != nil {
		cmd.Print("Error: ")
		cmd.PrintErr(e)
		cmd.Print("\n")
		os.Exit(-1)
	}
}

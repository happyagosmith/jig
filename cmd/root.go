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
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
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
}

func CheckErr(e error) {
	if e != nil {
		rootCmd.PrintErr(e)
		os.Exit(-1)
	}
}

func Execute(version string) {
	rootCmd.Flags().BoolP("version", "v", false, "version for myapp")
	rootCmd.Run = func(cmd *cobra.Command, args []string) {
		if v, _ := cmd.Flags().GetBool("version"); v {
			fmt.Println("jig version " + version)
			return
		}

		rootCmd.Help()
	}
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.jig.yaml)")
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.PersistentFlags().String("issuePattern", `(^j_(?P<jira_1>.*)$)|(?P<jira_2>^[^\_]+$)`, "Pattern to apply on the git commit message to extract the issue keys from the message. The pattern should include the named groups composed by noun with number (e.g. jira_1). The noun refers to the issue tracker (at the moment only jira is supported). The number has the purpose to define more than one pattern for the same issue tracker (this is usefull if the commit message format is changed over the time). The pattern must be a valid regex pattern.")
	viper.BindPFlag("issuePattern", rootCmd.PersistentFlags().Lookup("issuePattern"))

	rootCmd.PersistentFlags().Bool("withCCWithoutScope", false, "if true, extract conventional commit without scope")
	viper.BindPFlag("withCCWithoutScope", rootCmd.PersistentFlags().Lookup("withCCWithoutScope"))

	rootCmd.PersistentFlags().String("customCommitPattern", `\[(?P<scope>[^\]]*)\](?P<subject>.*)`, "Custom pattern to apply on the git commit message to extract the issue keys and the summary. If the message is not a conventional commit message, this custom pattern is applied. The pattern should include the named groups scope and subject")
	viper.BindPFlag("customCommitPattern", rootCmd.PersistentFlags().Lookup("customCommitPattern"))

	rootCmd.PersistentFlags().String("gitURL", "", "Git base URL")
	viper.BindPFlag("gitURL", rootCmd.PersistentFlags().Lookup("gitURL"))

	rootCmd.PersistentFlags().String("gitToken", "", "Git token with read REST API permissions")
	viper.BindPFlag("gitToken", rootCmd.PersistentFlags().Lookup("gitToken"))

	rootCmd.PersistentFlags().String("jiraURL", "", "Jira base URL")
	viper.BindPFlag("jiraURL", rootCmd.PersistentFlags().Lookup("jiraURL"))

	rootCmd.PersistentFlags().String("jiraUsername", "", "Jira username with read REST API permissions")
	viper.BindPFlag("jiraUsername", rootCmd.PersistentFlags().Lookup("jiraUsername"))

	rootCmd.PersistentFlags().String("jiraPassword", "", "Jira password/token with read REST API permissions")
	viper.BindPFlag("jiraPassword", rootCmd.PersistentFlags().Lookup("jiraPassword"))

	rootCmd.PersistentFlags().String("jiraClosedFeatureFilter", "Story:GOLIVE,TECH TASK:Completata", "List of filters type:status that identify the closed features")
	viper.BindPFlag("jiraClosedFeatureFilter", rootCmd.PersistentFlags().Lookup("jiraClosedFeatureFilter"))

	rootCmd.PersistentFlags().String("jiraFixedBugFilter", "BUG:FIXED,BUG:RELEASED", "List of filters type:status that identify the fixed bugs")
	viper.BindPFlag("jiraFixedBugFilter", rootCmd.PersistentFlags().Lookup("jiraFixedBugFilter"))

	rootCmd.PersistentFlags().String("jiraKnownIssuesJQL", "status not in (Done, RELEASED, Fixed, GOLIVE, Cancelled) AND issuetype in (Bug, \"TECH DEBT\")", "Jira JQL to retrieve the known issues")
	viper.BindPFlag("jiraKnownIssuesJQL", rootCmd.PersistentFlags().Lookup("jiraKnownIssuesJQL"))

	rootCmd.AddCommand(newEnrichCmd())
	rootCmd.AddCommand(newGenerateCmd())
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

/*
Copyright © 2023 Happy Smith happyagosmith@gmail.com
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	releaseNote "github.com/happyagosmith/jig/internal/releasenote"
)

func newGenerateCmd() *cobra.Command {
	generateCmd := &cobra.Command{
		Use:   "generate [template] -m [model.yaml]",
		Short: "generate [template] -m [model.yaml]",
		Long:  `Render the template using the values of the model.yaml file
		
In case --withEnrich is used, before rendering the template, jig executes the enrichment of the model in memory with
the data extracted from Git and Jira. Refer to the help of the enrich subcommand for details.`,
		Args: func(cmd *cobra.Command, args []string) error {
			return cobra.ExactArgs(1)(cmd, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			modelPath := viper.GetString("model")
			tplPath := args[0]
			cmd.Printf("Using model file: %s\n", modelPath)
			cmd.Printf("Using template file: %s\n", tplPath)

			v, err := os.ReadFile(modelPath)
			CheckErr(err)

			if enrich := viper.GetBool("withEnrich"); enrich {
				v = EnrichModel(v)
			}

			err = releaseNote.Generate(tplPath, v)
			CheckErr(err)

			return nil
		},
	}
	generateCmd.Flags().StringP("model", "m", "", "Path of the release notes model")
	viper.BindPFlag("model", generateCmd.Flags().Lookup("model"))
	generateCmd.Flags().Bool("withEnrich", false, "If true, enrich the model before generate")
	viper.BindPFlag("withEnrich", generateCmd.Flags().Lookup("withEnrich"))
	setEnrichFlags(generateCmd)
	return generateCmd
}

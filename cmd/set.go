/*
Copyright © 2023 Happy Smith happyagosmith@gmail.com
*/
package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/happyagosmith/jig/internal/model"
)

func setVersions() *cobra.Command {
	updateCmd := &cobra.Command{
		Use:   "setVersions",
		Short: "setVersions [model.yaml]",
		Long:  `The command is designed to automate the process of updating version information 
in the [model.yaml] file. 

This command works in conjunction with the checkVersion field in your model. The checkVersion 
field should specify the file and the YAML path from which the version information should be 
read. An example configuration might look like '@filepath:$.a.b'.

When you run jig setVersions, the command reads the version information from the specified file 
and YAML path and updates the version field in your model with this information. 

If the version field is updated with a new value, the previousVersion field is also updated. 
The previousVersion field will hold the value that was previously in the version field before 
the update.

This command is particularly useful for keeping your version information up to date without 
having to manually change the version and previousVersion fields each time a new version 
is released.`,
		Args: func(cmd *cobra.Command, args []string) error {
			return cobra.ExactArgs(1)(cmd, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			modelPath := args[0]

			fl := NewFileLoader(viper.GetString("gitToken"))
			cmd.Printf("using model file: %s\n", modelPath)
			b, err := fl.GetFile(modelPath)
			CheckErr(err)

			m, err := model.New(b)
			CheckErr(err)

			err = m.SetVersions(filepath.Dir(modelPath))
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

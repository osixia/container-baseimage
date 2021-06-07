package environment

import "github.com/spf13/cobra"

var EnvironmentCmd = &cobra.Command{
	Use:   "environment",
	Short: "Print, export, and save environment variables",

	Aliases: []string{
		"env",
	},
}

func init() {
	//subcommands
	EnvironmentCmd.AddCommand(filesCmd)
	EnvironmentCmd.AddCommand(printCmd)
}

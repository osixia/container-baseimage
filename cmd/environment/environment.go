package environment

import "github.com/spf13/cobra"

var EnvironmentCmd = &cobra.Command{
	Use:   "environment",
	Short: "Show environment variables and environment files",

	Aliases: []string{
		"env",
	},
}

func init() {
	//subcommands
	EnvironmentCmd.AddCommand(printCmd)
	EnvironmentCmd.AddCommand(filesCmd)
}

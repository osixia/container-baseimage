package provider

import "github.com/spf13/cobra"

var ProviderCmd = &cobra.Command{
	Use:   "provider",
	Short: "Providers",

	Aliases: []string{
		"p",
	},
}

func init() {
	//subcommands
	ProviderCmd.AddCommand(githubCmd)
}

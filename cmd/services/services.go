package services

import "github.com/spf13/cobra"

var ServicesCmd = &cobra.Command{
	Use:   "services",
	Short: "Services subcommands",

	Aliases: []string{
		"s",
	},
}

func init() {
	//subcommands
	ServicesCmd.AddCommand(requireCmd)
	ServicesCmd.AddCommand(installCmd)
}

package services

import "github.com/spf13/cobra"

var ServicesCmd = &cobra.Command{
	Use:   "services",
	Short: "Services subcommands",

	Aliases: []string{
		"svcs",
	},
}

func init() {
	//subcommands
	ServicesCmd.AddCommand(requireCmd)
	ServicesCmd.AddCommand(downloadCmd)
	ServicesCmd.AddCommand(installCmd)
	ServicesCmd.AddCommand(linkCmd)
	ServicesCmd.AddCommand(unlinkCmd)
	ServicesCmd.AddCommand(statusCmd)
}

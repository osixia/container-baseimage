package services

import "github.com/spf13/cobra"

var TagNamePrefix = "tag:"

var ServicesCmd = &cobra.Command{
	Use:   "services",
	Short: "Manage services",

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

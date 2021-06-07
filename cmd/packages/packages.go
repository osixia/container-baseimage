package packages

import "github.com/spf13/cobra"

var PackagesCmd = &cobra.Command{
	Use:   "packages",
	Short: "Manage packages",

	Aliases: []string{
		"pkgs",
	},
}

func init() {
	//subcommands
	PackagesCmd.AddCommand(installCmd)
	PackagesCmd.AddCommand(updateCmd)
	PackagesCmd.AddCommand(cleanCmd)
	PackagesCmd.AddCommand(removeCmd)
}

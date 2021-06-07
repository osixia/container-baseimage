package packages

import "github.com/spf13/cobra"

var PackagesCmd = &cobra.Command{
	Use:   "packages",
	Short: "Packages subcommands",

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

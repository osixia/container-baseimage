package groups

import "github.com/spf13/cobra"

var GroupsCmd = &cobra.Command{
	Use:   "groups",
	Short: "Manage groups",

	Aliases: []string{
		"g",
	},
}

func init() {
	//subcommands
	GroupsCmd.AddCommand(addCmd)
}

package processes

import "github.com/spf13/cobra"

var TagNamePrefix = "tag:"

var ProcessesCmd = &cobra.Command{
	Use:   "processes",
	Short: "Processes subcommands",

	Aliases: []string{
		"prcs",
	},
}

func init() {
	//subcommands
	ProcessesCmd.AddCommand(startCmd)
	ProcessesCmd.AddCommand(stopCmd)
	ProcessesCmd.AddCommand(statusCmd)
}

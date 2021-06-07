package job

import (
	"github.com/spf13/cobra"
)

var JobCmd = &cobra.Command{
	Use:   "job",
	Short: "Jobs",

	Aliases: []string{
		"j",
	},
}

func init() {
	//subcommands
	JobCmd.AddCommand(buildCmd)
	JobCmd.AddCommand(exportCmd)
	JobCmd.AddCommand(testCmd)
	JobCmd.AddCommand(deployCmd)
	JobCmd.AddCommand(publishCmd)
}

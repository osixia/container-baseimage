package generate

import (
	"github.com/spf13/cobra"

	"github.com/osixia/container-baseimage/cmd/logger"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/log"
)

var environmentCmdFlags = &generateFlags{}

var environmentCmd = &cobra.Command{
	Use:   "environment",
	Short: "Generate environment",

	Aliases: []string{
		"e",
	},

	Args: cobra.NoArgs,

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		files, err := core.Instance().Generator().GenerateEnvironment()
		if err != nil {
			log.Fatalf("%v: %v", cmd.Use, err.Error())
		}

		if environmentCmdFlags.print {
			print(files)
		}
	},
}

func init() {
	// flags
	environmentCmd.Flags().SortFlags = false
	addGenerateFlags(environmentCmd.Flags(), environmentCmdFlags)
	logger.AddFlags(environmentCmd.Flags())
}

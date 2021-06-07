package packages

import (
	"github.com/spf13/cobra"

	"github.com/osixia/container-baseimage/cmd/logger"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/log"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean packages cache",

	Aliases: []string{
		"c",
	},

	Args: cobra.NoArgs,

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		if err := core.Instance().Distribution().CleanPackages(cmd.Context()); err != nil {
			log.Fatalf("%v: %v", cmd.Use, err.Error())
		}
	},
}

func init() {
	// flags
	cleanCmd.Flags().SortFlags = false
	logger.AddFlags(cleanCmd.Flags())
}

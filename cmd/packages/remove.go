package packages

import (
	"github.com/spf13/cobra"

	"github.com/osixia/container-baseimage/cmd/logger"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/log"
)

var removeCmd = &cobra.Command{
	Use:   "remove package [package]...",
	Short: "Remove package(s)",

	Args: cobra.MinimumNArgs(1),

	Aliases: []string{
		"r",
	},

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		if err := core.Instance().Distribution().RemovePackages(cmd.Context(), args); err != nil {
			log.Fatalf("%v: %v", cmd.Use, err.Error())
		}
	},
}

func init() {
	// flags
	removeCmd.Flags().SortFlags = false
	logger.AddFlags(removeCmd.Flags())
}

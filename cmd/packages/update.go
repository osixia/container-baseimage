package packages

import (
	"github.com/spf13/cobra"

	"github.com/osixia/container-baseimage/cmd/logger"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/log"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update list of available packages",

	Args: cobra.NoArgs,

	Aliases: []string{
		"u",
	},

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		if err := core.Instance().Distribution().UpdatePackages(cmd.Context()); err != nil {
			log.Fatalf("%v: %v", cmd.Use, err.Error())
		}
	},
}

func init() {
	// flags
	updateCmd.Flags().SortFlags = false
	logger.AddFlags(updateCmd.Flags())
}

package packages

import (
	"github.com/spf13/cobra"

	cmdlog "github.com/osixia/container-baseimage/cmd/log"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/log"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update list of available packages",

	Aliases: []string{
		"u",
	},

	Args: cobra.NoArgs,

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
	cmdlog.AddFlags(updateCmd.Flags())
}

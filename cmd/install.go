package cmd

import (
	"github.com/spf13/cobra"

	cmdlog "github.com/osixia/container-baseimage/cmd/log"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/helpers"
	"github.com/osixia/container-baseimage/log"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install the container binary and prepare the runtime",

	GroupID: mainGroupID,

	Aliases: []string{
		"inst",
	},

	Args: cobra.NoArgs,

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		helpers.Must(core.Instance().Install(cmd.Context()))
	},
}

func init() {
	// flags
	installCmd.Flags().SortFlags = false
	cmdlog.AddFlags(installCmd.Flags())
}

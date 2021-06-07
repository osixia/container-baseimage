package cmd

import (
	"github.com/spf13/cobra"

	cmdlog "github.com/osixia/container-baseimage/cmd/log"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/log"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install the container binary and prepare the runtime",

	GroupID: mainGroupID,

	Args: cobra.NoArgs,

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		if err := core.Instance().Install(cmd.Context()); err != nil {
			log.Fatalf("%v: %v", cmd.Use, err.Error())
		}
	},
}

func init() {
	// flags
	installCmd.Flags().SortFlags = false
	cmdlog.AddFlags(installCmd.Flags())
}

package cmd

import (
	"github.com/spf13/cobra"

	"github.com/osixia/container-baseimage/cmd/logger"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/log"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install container-baseimage",

	GroupID: installGroupID,

	Aliases: []string{
		"i",
	},

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
	logger.AddFlags(installCmd.Flags())
}

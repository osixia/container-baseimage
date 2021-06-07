package cmd

import (
	"github.com/spf13/cobra"

	"github.com/osixia/container-baseimage/cmd/logger"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/log"
)

var debugCmd = &cobra.Command{
	Use:   "debug [package]...",
	Short: "Install debug utilities",

	GroupID: installGroupID,

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		pkgs := append(core.Instance().Distribution().Config().DebugPackages, args...)

		if err := core.Instance().Distribution().InstallPackages(cmd.Context(), pkgs, true, false); err != nil {
			log.Fatalf("%v: %v", cmd.Use, err.Error())
		}
	},
}

func init() {
	// flags
	debugCmd.Flags().SortFlags = false
	logger.AddFlags(debugCmd.Flags())
}

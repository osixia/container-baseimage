package cmd

import (
	"github.com/spf13/cobra"

	cmdlog "github.com/osixia/container-baseimage/cmd/log"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/log"
)

var debugCmd = &cobra.Command{
	Use:   "debug [package]...",
	Short: "Install debug utilities",

	GroupID: setupGroupID,

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
	cmdlog.AddFlags(debugCmd.Flags())
}

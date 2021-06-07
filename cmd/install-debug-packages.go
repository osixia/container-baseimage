package cmd

import (
	"github.com/spf13/cobra"

	"github.com/osixia/container-baseimage/cmd/logger"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/log"
)

var installDebugPackagesCmd = &cobra.Command{
	Use:   "install-debug-packages [extra_package]...",
	Short: "Install debug packages",

	GroupID: distributionGroupID,

	Aliases: []string{
		"idp",
	},

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		log.Tracef("%v", core.Instance().Distribution())

		debugPackages := core.Instance().Distribution().Config().DebugPackages
		packages := append(debugPackages, args...)

		if err := core.Instance().Distribution().InstallPackages(cmd.Context(), packages); err != nil {
			log.Fatalf("%v: %v", cmd.Use, err.Error())
		}
	},
}

func init() {
	// flags
	installDebugPackagesCmd.Flags().SortFlags = false
	logger.AddFlags(installDebugPackagesCmd.Flags())
}

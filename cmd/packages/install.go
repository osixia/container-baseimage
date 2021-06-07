package packages

import (
	"github.com/spf13/cobra"

	"github.com/osixia/container-baseimage/cmd/logger"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/log"
)

var installCmdFlags = &installFlags{}

type installFlags struct {
	update bool
	clean  bool
}

var installCmd = &cobra.Command{
	Use:   "install package [package]...",
	Short: "Install packages",

	Args: cobra.MinimumNArgs(1),

	Aliases: []string{
		"i",
	},

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		if err := core.Instance().Distribution().InstallPackages(cmd.Context(), args, installCmdFlags.update, installCmdFlags.clean); err != nil {
			log.Fatalf("%v: %v", cmd.Use, err.Error())
		}
	},
}

func init() {
	// flags
	installCmd.Flags().SortFlags = false

	installCmd.Flags().BoolVarP(&installCmdFlags.update, "update", "u", false, "update list of available packages before install")
	installCmd.Flags().BoolVarP(&installCmdFlags.clean, "clean", "c", false, "clean packages cache after install")

	logger.AddFlags(installCmd.Flags())
}

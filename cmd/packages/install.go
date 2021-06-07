package packages

import (
	"github.com/spf13/cobra"

	cmdlog "github.com/osixia/container-baseimage/cmd/log"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/helpers"
	"github.com/osixia/container-baseimage/log"
)

var installCmdFlags = &installFlags{}

type installFlags struct {
	update bool
	clean  bool

	debug bool
}

var installCmd = &cobra.Command{
	Use:   "install package [package]...",
	Short: "Install packages",

	Aliases: []string{
		"inst",
	},

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		pkgs := args

		if installCmdFlags.debug {
			pkgs = append(core.Instance().Distribution().Config().DebugPackages, pkgs...)
		}

		helpers.Must(core.Instance().Distribution().InstallPackages(cmd.Context(), pkgs, installCmdFlags.update, installCmdFlags.clean))
	},
}

func init() {
	// flags
	installCmd.Flags().SortFlags = false

	installCmd.Flags().BoolVarP(&installCmdFlags.update, "update", "u", false, "update list of available packages before install")
	installCmd.Flags().BoolVarP(&installCmdFlags.clean, "clean", "c", false, "clean packages cache after install\n")

	installCmd.Flags().BoolVarP(&installCmdFlags.debug, "debug-tools", "d", false, "include common debugging tools\n")

	cmdlog.AddFlags(installCmd.Flags())
}

package services

import (
	"fmt"

	"github.com/spf13/cobra"

	cmdlog "github.com/osixia/container-baseimage/cmd/log"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/helpers"
	"github.com/osixia/container-baseimage/log"
)

var installCmd = &cobra.Command{
	Use:   fmt.Sprintf("install [service|%vname]...", TagNamePrefix),
	Short: "Install services",

	Long: "Install services\nWith no argument: install all not optional services and only downloaded optional services",

	Aliases: []string{
		"inst",
	},

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		ss := helpers.MustVal(core.Instance().Services().List(core.WithServicesNames(args), core.HandleServicesTagPrefixInNames(TagNamePrefix)))

		// if no service defined
		if len(ss) == 0 {

			// get all not optional services that are not already installed
			s1 := helpers.MustVal(core.Instance().Services().List(core.WithServicesOptional(false), core.WithServicesInstalled(false)))
			ss = append(ss, s1...)

			// get all downloaded optional services that are not already installed
			s2 := helpers.MustVal(core.Instance().Services().List(core.WithServicesOptional(true), core.WithServicesDownloaded(true), core.WithServicesInstalled(false)))
			ss = append(ss, s2...)

			core.Instance().Services().SortByPriority(ss)
		}

		helpers.Must(core.Instance().Services().Install(cmd.Context(), ss))
	},
}

func init() {
	// flags
	installCmd.Flags().SortFlags = false
	cmdlog.AddFlags(installCmd.Flags())
}

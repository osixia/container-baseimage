package services

import (
	"fmt"

	"github.com/spf13/cobra"

	cmdlog "github.com/osixia/container-baseimage/cmd/log"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/helpers"
	"github.com/osixia/container-baseimage/log"
)

var downloadCmd = &cobra.Command{
	Use:   fmt.Sprintf("download [service|%vname]...", TagNamePrefix),
	Short: "Download services",

	Long: "Download services\nWith no argument: download all not optional services",

	Aliases: []string{
		"dl",
	},

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		ss := helpers.MustVal(core.Instance().Services().List(core.WithServicesNames(args), core.HandleServicesTagPrefixInNames(TagNamePrefix)))

		// if no service defined: get all not optional services that are not already downloaded
		if len(ss) == 0 {
			ss = helpers.MustVal(core.Instance().Services().List(core.WithServicesOptional(false), core.WithServicesDownloaded(false), core.SortServicesByPriority(true)))
		}

		helpers.Must(core.Instance().Services().Download(cmd.Context(), ss))
	},
}

func init() {
	// flags
	downloadCmd.Flags().SortFlags = false
	cmdlog.AddFlags(downloadCmd.Flags())
}

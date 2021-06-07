package services

import (
	"fmt"

	"github.com/spf13/cobra"

	cmdlog "github.com/osixia/container-baseimage/cmd/log"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/helpers"
	"github.com/osixia/container-baseimage/log"
)

var unlinkCmd = &cobra.Command{
	Use:   fmt.Sprintf("unlink [service|%vname]...", TagNamePrefix),
	Short: "Unlink entrypoint's services",

	Long: "Unlink entrypoint's services\nWith no argument: unlink all linked services",

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		ss := helpers.MustVal(core.Instance().Services().List(core.WithServicesNames(args), core.HandleServicesTagPrefixInNames(TagNamePrefix)))

		// if no service defined: get all linked services
		if len(ss) == 0 {
			ss = helpers.MustVal(core.Instance().Services().List(core.WithServicesLinked(true), core.SortServicesByPriority(true)))
		}

		helpers.Must(core.Instance().Services().Unlink(ss))
	},
}

func init() {
	// flags
	unlinkCmd.Flags().SortFlags = false
	cmdlog.AddFlags(unlinkCmd.Flags())
}

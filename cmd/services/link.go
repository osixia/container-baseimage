package services

import (
	"fmt"

	"github.com/spf13/cobra"

	cmdlog "github.com/osixia/container-baseimage/cmd/log"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/helpers"
	"github.com/osixia/container-baseimage/log"
)

var linkCmd = &cobra.Command{
	Use:   fmt.Sprintf("link [service|%vname]...", TagNamePrefix),
	Short: "Link services to entrypoint",

	Long: "Link services to entrypoint\nWith no argument: link all not optional services",

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		ss := helpers.MustVal(core.Instance().Services().List(core.WithServicesNames(args), core.HandleServicesTagPrefixInNames(TagNamePrefix)))

		// if no service defined: get all not optional services that are not already linked
		if len(ss) == 0 {
			ss = helpers.MustVal(core.Instance().Services().List(core.WithServicesOptional(false), core.WithServicesLinked(false), core.SortServicesByPriority(true)))
		}

		helpers.Must(core.Instance().Services().Link(ss))
	},
}

func init() {
	// flags
	linkCmd.Flags().SortFlags = false
	cmdlog.AddFlags(linkCmd.Flags())
}

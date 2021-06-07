package services

import (
	"fmt"

	"github.com/spf13/cobra"

	cmdlog "github.com/osixia/container-baseimage/cmd/log"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/helpers"
	"github.com/osixia/container-baseimage/log"
)

var requireCmd = &cobra.Command{
	Use:   fmt.Sprintf("require service|%vname [service|%vname]...", TagNamePrefix, TagNamePrefix),
	Short: "Require optional service",

	Aliases: []string{
		"req",
	},

	Args: cobra.MinimumNArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		ss := helpers.MustVal(core.Instance().Services().List(core.WithServicesNames(args), core.HandleServicesTagPrefixInNames(TagNamePrefix)))

		helpers.Must(core.Instance().Services().Require(ss))
	},
}

func init() {
	// flags
	requireCmd.Flags().SortFlags = false
	cmdlog.AddFlags(requireCmd.Flags())
}

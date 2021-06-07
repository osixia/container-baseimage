package services

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/osixia/container-baseimage/cmd/logger"
	"github.com/osixia/container-baseimage/config"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/log"
)

var requireCmd = &cobra.Command{
	Use:   fmt.Sprintf("require service|%vname [service|%vname]...", config.TagsNamePrefix, config.TagsNamePrefix),
	Short: "Require optional service",

	Aliases: []string{
		"r",
	},

	Args: cobra.MinimumNArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		ss, err := core.Instance().Services().List(core.WithServicesNames(args), core.HandleServicesTagPrefixInNames(config.TagsNamePrefix))
		if err != nil {
			log.Fatalf("%v: %v", cmd.Use, err.Error())
		}

		if err := core.Instance().Services().Require(ss); err != nil {
			log.Fatalf("%v: %v", cmd.Use, err.Error())
		}
	},
}

func init() {
	// flags
	requireCmd.Flags().SortFlags = false
	logger.AddFlags(requireCmd.Flags())
}

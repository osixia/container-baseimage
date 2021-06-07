package services

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/osixia/container-baseimage/cmd/logger"
	"github.com/osixia/container-baseimage/config"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/log"
)

var installCmd = &cobra.Command{
	Use:   fmt.Sprintf("install [service|%vname]...", config.TagsNamePrefix),
	Short: "Install service(s)",

	Long: "With no argument: install all not optional services and only downloaded optional services",

	Aliases: []string{
		"i",
	},

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		ss, err := core.Instance().Services().List(core.WithServicesNames(args), core.HandleServicesTagPrefixInNames(config.TagsNamePrefix))
		if err != nil {
			log.Fatalf("%v: %v", cmd.Use, err.Error())
		}

		// if no service defined install all not optional services not already installed
		if len(ss) == 0 {
			ss, err = core.Instance().Services().List(core.WithServicesOptional(false), core.WithServicesInstalled(false), core.SortServicesByPriority(true))
			if err != nil {
				log.Fatalf("%v: %v", cmd.Use, err.Error())
			}
		}

		if err := core.Instance().Services().Install(cmd.Context(), ss); err != nil {
			log.Fatalf("%v: %v", cmd.Use, err.Error())
		}
	},
}

func init() {
	// flags
	installCmd.Flags().SortFlags = false
	logger.AddFlags(installCmd.Flags())
}

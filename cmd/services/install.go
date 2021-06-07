package services

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/osixia/container-baseimage/cmd/logger"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/log"
)

var installCmd = &cobra.Command{
	Use:   fmt.Sprintf("install [service|%vname]...", TagNamePrefix),
	Short: "Install services",

	Long: "Install services\nWith no argument: install all not optional services and only downloaded optional services",

	Aliases: []string{
		"i",
	},

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		ss, err := core.Instance().Services().List(core.WithServicesNames(args), core.HandleServicesTagPrefixInNames(TagNamePrefix))
		if err != nil {
			log.Fatalf("%v: %v", cmd.Use, err.Error())
		}

		// if no service defined
		if len(ss) == 0 {

			// get all not optional services that are not already installed
			if s1, err := core.Instance().Services().List(core.WithServicesOptional(false), core.WithServicesInstalled(false)); err != nil {
				log.Fatalf("%v: %v", cmd.Use, err.Error())
			} else {
				ss = append(ss, s1...)
			}

			// get all downloaded optional services that are not already installed
			if s2, err := core.Instance().Services().List(core.WithServicesOptional(true), core.WithServicesDownloaded(true), core.WithServicesInstalled(false)); err != nil {
				log.Fatalf("%v: %v", cmd.Use, err.Error())
			} else {
				ss = append(ss, s2...)
			}

			core.Instance().Services().SortByPriority(ss)
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

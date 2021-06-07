package services

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/osixia/container-baseimage/cmd/logger"
	"github.com/osixia/container-baseimage/config"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/log"
)

var downloadCmd = &cobra.Command{
	Use:   fmt.Sprintf("download [service|%vname]...", config.TagsNamePrefix),
	Short: "Download optional service",

	Long: "With no argument: download all not optional services",

	Aliases: []string{
		"d",
	},

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		ss, err := core.Instance().Services().List(core.WithServicesNames(args), core.HandleServicesTagPrefixInNames(config.TagsNamePrefix))
		if err != nil {
			log.Fatalf("%v: %v", cmd.Use, err.Error())
		}

		// if no service defined download all not optional and not already downloaded services
		if len(ss) == 0 {
			ss, err = core.Instance().Services().List(core.WithServicesOptional(false), core.WithServicesDownloaded(false), core.SortServicesByPriority(true))
			if err != nil {
				log.Fatalf("%v: %v", cmd.Use, err.Error())
			}
		}

		if err := core.Instance().Services().Download(cmd.Context(), ss); err != nil {
			log.Fatalf("%v: %v", cmd.Use, err.Error())
		}
	},
}

func init() {
	// flags
	downloadCmd.Flags().SortFlags = false
	logger.AddFlags(downloadCmd.Flags())
}

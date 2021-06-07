package services

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/osixia/container-baseimage/cmd/logger"
	"github.com/osixia/container-baseimage/config"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/log"
)

var unlinkCmd = &cobra.Command{
	Use:   fmt.Sprintf("unlink [service|%vname]...", config.TagsNamePrefix),
	Short: "Unlink entrypoint's service(s)",

	Long: "With no argument: unlink all linked services",

	Aliases: []string{
		"u",
	},

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		ss, err := core.Instance().Services().List(core.WithServicesNames(args), core.HandleServicesTagPrefixInNames(config.TagsNamePrefix))
		if err != nil {
			log.Fatalf("%v: %v", cmd.Use, err.Error())
		}

		// if no service defined unlink all linked services
		if len(ss) == 0 {
			ss, err = core.Instance().Services().List(core.WithServicesLinked(true), core.SortServicesByPriority(true))
			if err != nil {
				log.Fatalf("%v: %v", cmd.Use, err.Error())
			}
		}

		if err := core.Instance().Services().Unlink(ss); err != nil {
			log.Fatalf("%v: %v", cmd.Use, err.Error())
		}
	},
}

func init() {
	// flags
	unlinkCmd.Flags().SortFlags = false
	logger.AddFlags(unlinkCmd.Flags())
}

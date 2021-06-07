package services

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/osixia/container-baseimage/cmd/logger"
	"github.com/osixia/container-baseimage/config"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/log"
)

var linkCmd = &cobra.Command{
	Use:   fmt.Sprintf("link [service|%vname]...", config.TagsNamePrefix),
	Short: "Link service(s) to entrypoint",

	Long: "With no argument: link all installed and not optional services",

	Aliases: []string{
		"l",
	},

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		ss, err := core.Instance().Services().List(core.WithServicesNames(args), core.HandleServicesTagPrefixInNames(config.TagsNamePrefix))
		if err != nil {
			log.Fatalf("%v: %v", cmd.Use, err.Error())
		}

		// if no service defined link all installed, not optional and not already linked services
		if len(ss) == 0 {
			ss, err = core.Instance().Services().List(core.WithServicesOptional(false), core.WithServicesLinked(false), core.SortServicesByPriority(true))
			if err != nil {
				log.Fatalf("%v: %v", cmd.Use, err.Error())
			}
		}

		if err := core.Instance().Services().Link(ss); err != nil {
			log.Fatalf("%v: %v", cmd.Use, err.Error())
		}
	},
}

func init() {
	// flags
	linkCmd.Flags().SortFlags = false
	logger.AddFlags(linkCmd.Flags())
}

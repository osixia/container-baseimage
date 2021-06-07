package services

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/osixia/container-baseimage/cmd/logger"
	"github.com/osixia/container-baseimage/config"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/log"
)

var statusCmd = &cobra.Command{
	Use:   fmt.Sprintf("status [service|%vname]...", config.TagsNamePrefix),
	Short: "Print service(s) status",

	Aliases: []string{
		"s",
	},

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		ss, err := core.Instance().Services().List(core.WithServicesNames(args), core.HandleServicesTagPrefixInNames(config.TagsNamePrefix))
		if err != nil {
			log.Fatalf("%v: %v", cmd.Use, err.Error())
		}

		// if no process defined get all processes
		if len(ss) == 0 {
			ss, err = core.Instance().Services().List()
			if err != nil {
				log.Fatalf("%v: %v", cmd.Use, err.Error())
			}
		}

		if len(ss) == 0 {
			fmt.Println("No services found")
			return
		}

		for _, p := range ss {
			fmt.Println(p.Status())
		}
	},
}

func init() {
	// flags
	statusCmd.Flags().SortFlags = false
	logger.AddFlags(statusCmd.Flags())
}

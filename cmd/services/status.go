package services

import (
	"fmt"

	"github.com/spf13/cobra"

	cmdlog "github.com/osixia/container-baseimage/cmd/log"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/log"
)

var statusCmd = &cobra.Command{
	Use:   fmt.Sprintf("status [service|%vname]...", TagNamePrefix),
	Short: "Print services status",

	Aliases: []string{
		"s",
	},

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		ss, err := core.Instance().Services().List(core.WithServicesNames(args), core.HandleServicesTagPrefixInNames(TagNamePrefix))
		if err != nil {
			log.Fatalf("%v: %v", cmd.Use, err.Error())
		}

		// if no service defined: get all services
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

		for _, s := range ss {
			fmt.Println(s.Status())
		}
	},
}

func init() {
	// flags
	statusCmd.Flags().SortFlags = false
	cmdlog.AddFlags(statusCmd.Flags())
}

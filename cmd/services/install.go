package services

import (
	"github.com/spf13/cobra"

	"github.com/osixia/container-baseimage/cmd/logger"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/log"
)

var installCmd = &cobra.Command{
	Use:   "install [service]...",
	Short: "Install service(s)",

	Aliases: []string{
		"i",
	},

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		services := make([]core.Service, 0, len(args))

		for _, name := range args {
			s, err := core.Instance().Services().Get(name)
			if err != nil {
				log.Fatalf("%v: %v", cmd.Use, err.Error())
			}

			services = append(services, s)
		}

		// if no service defined install all not optional and not installed services
		if len(services) == 0 {
			svcs, err := core.Instance().Services().List(core.WithOptionalServices(false), core.WithInstalledServices(false))
			if err != nil {
				log.Fatalf("%v: %v", cmd.Use, err.Error())
			}
			services = svcs
		}

		if err := core.Instance().Services().Install(cmd.Context(), services); err != nil {
			log.Fatalf("%v: %v", cmd.Use, err.Error())
		}
	},
}

func init() {
	// flags
	installCmd.Flags().SortFlags = false
	logger.AddFlags(installCmd.Flags())
}

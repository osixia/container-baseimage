package entrypoint

import (
	"github.com/spf13/cobra"

	"github.com/osixia/container-baseimage/cmd/logger"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/log"
)

var linkToEntrypointCmd = &cobra.Command{
	Use:   "link-services [service]...",
	Short: "Link service(s) to entrypoint",

	Aliases: []string{
		"ls",
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

		// if no service defined link all installed services
		if len(services) == 0 {
			svcs, err := core.Instance().Services().List(core.WithInstalledServices(true))
			if err != nil {
				log.Fatalf("%v: %v", cmd.Use, err.Error())
			}
			services = svcs
		}

		if _, err := core.Instance().Entrypoint().LinkServices(services); err != nil {
			log.Fatalf("%v: %v", cmd.Use, err.Error())
		}
	},
}

func init() {
	// flags
	linkToEntrypointCmd.Flags().SortFlags = false
	logger.AddFlags(linkToEntrypointCmd.Flags())
}

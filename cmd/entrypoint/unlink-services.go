package entrypoint

import (
	"github.com/spf13/cobra"

	"github.com/osixia/container-baseimage/cmd/logger"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/log"
)

var unlinkToEntrypointCmd = &cobra.Command{
	Use:   "unlink-services [service]...",
	Short: "Unlink entrypoint's service(s)",

	Aliases: []string{
		"us",
	},

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		linkedServices := make([]core.EntrypointLinkedService, 0, len(args))

		for _, name := range args {
			s, err := core.Instance().Entrypoint().GetLinkedService(name)
			if err != nil {
				log.Fatalf("%v: %v", cmd.Use, err.Error())
			}

			linkedServices = append(linkedServices, s)
		}

		// if no service defined unlink all services
		if len(linkedServices) == 0 {
			svcs, err := core.Instance().Entrypoint().ListLinkedServices()
			if err != nil {
				log.Fatalf("%v: %v", cmd.Use, err.Error())
			}
			linkedServices = svcs
		}

		if _, err := core.Instance().Entrypoint().UnlinkServices(linkedServices); err != nil {
			log.Fatalf("%v: %v", cmd.Use, err.Error())
		}
	},
}

func init() {
	// flags
	unlinkToEntrypointCmd.Flags().SortFlags = false
	logger.AddFlags(unlinkToEntrypointCmd.Flags())
}

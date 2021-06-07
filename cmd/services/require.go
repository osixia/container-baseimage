package services

import (
	"github.com/spf13/cobra"

	"github.com/osixia/container-baseimage/cmd/logger"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/log"
)

var requireCmd = &cobra.Command{
	Use:   "require service [name]...",
	Short: "Require optional service",

	Aliases: []string{
		"r",
	},

	Args: cobra.MinimumNArgs(1),

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

		if err := core.Instance().Services().Require(cmd.Context(), services); err != nil {
			log.Fatalf("%v: %v", cmd.Use, err.Error())
		}
	},
}

func init() {
	// flags
	requireCmd.Flags().SortFlags = false
	logger.AddFlags(requireCmd.Flags())
}

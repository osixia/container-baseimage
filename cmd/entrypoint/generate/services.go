package generate

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/osixia/container-baseimage/cmd/logger"
	"github.com/osixia/container-baseimage/config"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/log"
)

type generateServicesFlags struct {
	core.GenerateServicesOptions
	generateFlags
}

var servicesCmdFlags = &generateServicesFlags{}

var servicesCmd = &cobra.Command{
	Use:   "services [name]...",
	Short: "Generate services",

	Aliases: []string{
		"s",
	},

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		servicesCmdFlags.Names = args

		files, err := core.Instance().Generator().GenerateServices(&servicesCmdFlags.GenerateServicesOptions)
		if err != nil {
			log.Fatalf("%v: %v", cmd.Use, err.Error())
		}

		if servicesCmdFlags.print {
			print(files)
		}
	},
}

func init() {
	// flags
	servicesCmd.Flags().SortFlags = false
	addServicesFlags(servicesCmd.Flags(), &servicesCmdFlags.GenerateServicesOptions)
	addGenerateFlags(servicesCmd.Flags(), &servicesCmdFlags.generateFlags)
	logger.AddFlags(servicesCmd.Flags())
}

func addServicesFlags(fs *pflag.FlagSet, gopt *core.GenerateServicesOptions) {
	fs.IntVarP(&gopt.Priority, "priority", "p", config.ServicesConfig.DefaultPriority, "services priority")
	fs.BoolVar(&gopt.Optional, "optional", false, "optional service")
}

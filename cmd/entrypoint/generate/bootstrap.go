package generate

import (
	"github.com/spf13/cobra"

	"github.com/osixia/container-baseimage/cmd/logger"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/log"
)

type generateBootstrapFlags struct {
	core.GenerateBootstrapOptions
	generateFlags
}

var bootstrapCmdFlags = &generateBootstrapFlags{}

var bootstrapCmd = &cobra.Command{
	Use:   "bootstrap [service name]...",
	Short: "Generate bootstrap",

	Aliases: []string{
		"b",
	},

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		bootstrapCmdFlags.Names = args

		files, err := core.Instance().Generator().GenerateBootstrap(&bootstrapCmdFlags.GenerateBootstrapOptions)
		if err != nil {
			log.Fatalf("%v: %v", cmd.Use, err.Error())
		}

		if bootstrapCmdFlags.print {
			print(files)
		}
	},
}

func init() {
	// flags
	bootstrapCmd.Flags().SortFlags = false
	addDockerfileFlags(bootstrapCmd.Flags(), &bootstrapCmdFlags.GenerateDockerfileOptions)
	addServicesFlags(bootstrapCmd.Flags(), &bootstrapCmdFlags.GenerateServicesOptions)
	addGenerateFlags(bootstrapCmd.Flags(), &bootstrapCmdFlags.generateFlags)
	logger.AddFlags(bootstrapCmd.Flags())
}

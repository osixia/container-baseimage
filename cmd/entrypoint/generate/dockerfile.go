package generate

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/osixia/container-baseimage/cmd/logger"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/log"
)

type generateDockerfileFlags struct {
	core.GenerateDockerfileOptions
	generateFlags
}

var dockerfileCmdFlags = &generateDockerfileFlags{}

var dockerfileCmd = &cobra.Command{
	Use:   "dockerfile",
	Short: "Generate Dockerfile",

	Aliases: []string{
		"d",
	},

	Args: cobra.NoArgs,

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		files, err := core.Instance().Generator().GenerateDockerfile(&dockerfileCmdFlags.GenerateDockerfileOptions)
		if err != nil {
			log.Fatalf("%v: %v", cmd.Use, err.Error())
		}

		if dockerfileCmdFlags.print {
			print(files)
		}
	},
}

func init() {
	// flags
	dockerfileCmd.Flags().SortFlags = false
	addDockerfileFlags(dockerfileCmd.Flags(), &dockerfileCmdFlags.GenerateDockerfileOptions)
	addGenerateFlags(dockerfileCmd.Flags(), &dockerfileCmdFlags.generateFlags)
	logger.AddFlags(dockerfileCmd.Flags())
}

func addDockerfileFlags(fs *pflag.FlagSet, gopt *core.GenerateDockerfileOptions) {
	fs.BoolVarP(&gopt.Multiprocess, "multiprocess", "m", false, "generate multiprocess example")
}

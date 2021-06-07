package generate

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	cmdlog "github.com/osixia/container-baseimage/cmd/log"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/helpers"
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
		"df",
	},

	Args: cobra.NoArgs,

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		files := helpers.MustVal(core.Instance().Generator().GenerateDockerfile(&dockerfileCmdFlags.GenerateDockerfileOptions))

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
	cmdlog.AddFlags(dockerfileCmd.Flags())
}

func addDockerfileFlags(fs *pflag.FlagSet, gopt *core.GenerateDockerfileOptions) {
	fs.StringVarP(&gopt.Image, "image", "i", "osixia/baseimage-example:latest", "image name")
	fs.StringVarP(&gopt.FromImage, "from-image", "f", "", "overwrite from image\n")
}

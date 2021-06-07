package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/osixia/container-baseimage/cmd/logger"
	"github.com/osixia/container-baseimage/helpers"
	"github.com/osixia/container-baseimage/log"
)

var envsubstTemplatesFilesSuffix = ".template"

var envsubstTemplatesCmd = &cobra.Command{
	Use:   fmt.Sprintf("envsubst-templates templates_dir [output_dir=templates_dir] [templates_files_suffix=%v]", envsubstTemplatesFilesSuffix),
	Short: "Envsubst templates",

	GroupID: envGroupID,

	Args: cobra.RangeArgs(1, 3),

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		templatesDir := args[0]
		outputDir := templatesDir

		if len(args) > 1 {
			outputDir = args[1]

			if len(args) > 2 {
				envsubstTemplatesFilesSuffix = args[2]
			}
		}

		if _, err := helpers.EnvsubstTemplates(templatesDir, outputDir, envsubstTemplatesFilesSuffix); err != nil {
			log.Fatalf("%v: %v", cmd.Use, err.Error())
		}
	},
}

func init() {
	// flags
	envsubstTemplatesCmd.Flags().SortFlags = false
	logger.AddFlags(envsubstTemplatesCmd.Flags())
}

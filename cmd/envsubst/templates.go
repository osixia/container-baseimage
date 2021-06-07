package envsubst

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/osixia/container-baseimage/cmd/logger"
	"github.com/osixia/container-baseimage/helpers"
	"github.com/osixia/container-baseimage/log"
)

var templatesFilesSuffix = ".template"

var templatesCmd = &cobra.Command{
	Use:   fmt.Sprintf("templates templates_dir [output_dir=templates_dir] [templates_files_suffix=%v]", templatesFilesSuffix),
	Short: fmt.Sprintf("Render all %v files recursively in a directory", templatesFilesSuffix),

	Args: cobra.RangeArgs(1, 3),

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		templatesDir := args[0]
		outputDir := templatesDir

		if len(args) > 1 {
			outputDir = args[1]

			if len(args) > 2 {
				templatesFilesSuffix = args[2]
			}
		}

		if _, err := helpers.EnvsubstTemplates(templatesDir, outputDir, templatesFilesSuffix); err != nil {
			log.Fatalf("%v: %v", cmd.Use, err.Error())
		}
	},
}

func init() {
	// flags
	templatesCmd.Flags().SortFlags = false
	logger.AddFlags(templatesCmd.Flags())
}

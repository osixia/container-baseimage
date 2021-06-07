package envsubst

import (
	"fmt"

	"github.com/spf13/cobra"

	cmdlog "github.com/osixia/container-baseimage/cmd/log"
	"github.com/osixia/container-baseimage/helpers"
	"github.com/osixia/container-baseimage/log"
)

var TemplatesFilesSuffix = ".template"

var templatesCmd = &cobra.Command{
	Use:   fmt.Sprintf("templates templates_dir [output_dir=templates_dir] [templates_files_suffix=%v]", TemplatesFilesSuffix),
	Short: fmt.Sprintf("Render all %v files recursively in a directory", TemplatesFilesSuffix),

	Args: cobra.RangeArgs(1, 3),

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		templatesDir := args[0]
		outputDir := templatesDir
		filesSuffix := TemplatesFilesSuffix

		if len(args) > 1 {
			outputDir = args[1]

			if len(args) > 2 {
				filesSuffix = args[2]
			}
		}

		helpers.MustVal(helpers.EnvsubstTemplates(templatesDir, outputDir, filesSuffix))
	},
}

func init() {
	// flags
	templatesCmd.Flags().SortFlags = false
	cmdlog.AddFlags(templatesCmd.Flags())
}

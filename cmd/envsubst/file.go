package envsubst

import (
	"github.com/spf13/cobra"

	"github.com/osixia/container-baseimage/cmd/logger"
	"github.com/osixia/container-baseimage/helpers"
	"github.com/osixia/container-baseimage/log"
)

var fileCmd = &cobra.Command{
	Use:   "file input [output=input]",
	Short: "Render a single file using environment variables",

	Args: cobra.RangeArgs(1, 2),

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		input := args[0]
		output := input

		if len(args) > 1 {
			output = args[1]
		}

		if err := helpers.EnvsubstFile(input, output); err != nil {
			log.Fatalf("%v: %v", cmd.Use, err.Error())
		}
	},
}

func init() {
	// flags
	fileCmd.Flags().SortFlags = false
	logger.AddFlags(fileCmd.Flags())
}

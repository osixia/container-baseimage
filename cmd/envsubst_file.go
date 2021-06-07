package cmd

import (
	"github.com/spf13/cobra"

	"github.com/osixia/container-baseimage/cmd/logger"
	"github.com/osixia/container-baseimage/helpers"
	"github.com/osixia/container-baseimage/log"
)

var envsubstFileCmd = &cobra.Command{
	Use:   "envsubst-file input [output=input]",
	Short: "Envsubst file",

	GroupID: envGroupID,

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
	envsubstFileCmd.Flags().SortFlags = false
	logger.AddFlags(envsubstFileCmd.Flags())
}

package groups

import (
	"github.com/spf13/cobra"

	"github.com/osixia/container-baseimage/cmd/logger"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/log"
)

var addCmd = &cobra.Command{
	Use:   "add id name",
	Short: "Add group",

	Aliases: []string{
		"a",
	},

	Args: cobra.ExactArgs(2),

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		if err := core.Instance().Distribution().AddGroup(cmd.Context(), args[0], args[1]); err != nil {
			log.Fatalf("%v: %v", cmd.Use, err.Error())
		}
	},
}

func init() {
	// flags
	addCmd.Flags().SortFlags = false
	logger.AddFlags(addCmd.Flags())
}

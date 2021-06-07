package groups

import (
	"github.com/spf13/cobra"

	cmdlog "github.com/osixia/container-baseimage/cmd/log"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/helpers"
	"github.com/osixia/container-baseimage/log"
)

var addCmd = &cobra.Command{
	Use:   "add id name",
	Short: "Add group",

	Args: cobra.ExactArgs(2),

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		helpers.Must(core.Instance().Distribution().AddGroup(cmd.Context(), args[0], args[1]))
	},
}

func init() {
	// flags
	addCmd.Flags().SortFlags = false
	cmdlog.AddFlags(addCmd.Flags())
}

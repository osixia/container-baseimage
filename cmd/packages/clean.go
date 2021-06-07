package packages

import (
	"github.com/spf13/cobra"

	cmdlog "github.com/osixia/container-baseimage/cmd/log"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/helpers"
	"github.com/osixia/container-baseimage/log"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean packages cache",

	Args: cobra.NoArgs,

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		helpers.Must(core.Instance().Distribution().CleanPackages(cmd.Context()))
	},
}

func init() {
	// flags
	cleanCmd.Flags().SortFlags = false
	cmdlog.AddFlags(cleanCmd.Flags())
}

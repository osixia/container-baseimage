package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	cmdlog "github.com/osixia/container-baseimage/cmd/log"
	"github.com/osixia/container-baseimage/helpers"
	"github.com/osixia/container-baseimage/log"
)

type watchFlags struct {
	exec []string
	once bool
}

var watchCmdFlags = &watchFlags{}

var watchCmd = &cobra.Command{
	Use:   "watch file [file]...",
	Short: "Watch files",

	GroupID: filesystemGroupID,

	Args: cobra.MinimumNArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		// Go exits on SIGHUP, which the entrypoint forwards to services: catch it,
		// there is nothing to reload. Unlike signal.Ignore, executed scripts keep
		// the default SIGHUP behavior.
		hangup := make(chan os.Signal, 1)
		signal.Notify(hangup, syscall.SIGHUP)

		helpers.Must(helpers.Watch(cmd.Context(), args, watchCmdFlags.exec, watchCmdFlags.once))
	},
}

func init() {
	// flags
	watchCmd.Flags().SortFlags = false

	watchCmd.Flags().StringArrayVarP(&watchCmdFlags.exec, "exec", "x", nil, "execute the given script on file change")
	watchCmd.Flags().BoolVarP(&watchCmdFlags.once, "once", "1", false, "exit after the first detected change\n")

	cmdlog.AddFlags(watchCmd.Flags())
}

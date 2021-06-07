package processes

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	cmdlog "github.com/osixia/container-baseimage/cmd/log"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/helpers"
	"github.com/osixia/container-baseimage/log"
)

var stopCmd = &cobra.Command{
	Use:   fmt.Sprintf("stop [process|%vname]...", TagNamePrefix),
	Short: "Stop processes",

	Long: "Stop processes\nWith no argument: stop all processes",

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		ps := helpers.MustVal(core.Instance().Processes().List(core.WithProcessesNames(args), core.HandleProcessesTagPrefixInNames(TagNamePrefix)))

		// if no process defined: get all processes
		if len(ps) == 0 {
			ps = helpers.MustVal(core.Instance().Processes().List())
		}

		timeoutSecs := helpers.MustVal(cmd.Flags().GetInt("timeout"))

		helpers.Must(core.Instance().Processes().Stop(ps, time.Duration(timeoutSecs)*time.Second))
	},
}

func init() {
	// flags
	stopCmd.Flags().SortFlags = false

	stopCmd.Flags().Int("timeout", 30, "timeout in seconds to wait for process to be down (0 = no timeout)\n")

	cmdlog.AddFlags(stopCmd.Flags())
}

package processes

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/osixia/container-baseimage/cmd/logger"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/log"
)

var startCmd = &cobra.Command{
	Use:   fmt.Sprintf("start [process|%vname]...", TagNamePrefix),
	Short: "Start processes",

	Long: "Start processes\nWith no argument: start all processes",

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		ps, err := core.Instance().Processes().List(core.WithProcessesNames(args), core.HandleProcessesTagPrefixInNames(TagNamePrefix))
		if err != nil {
			log.Fatalf("%v: %v", cmd.Use, err.Error())
		}

		// if no process defined: get all processes
		if len(ps) == 0 {
			ps, err = core.Instance().Processes().List()
			if err != nil {
				log.Fatalf("%v: %v", cmd.Use, err.Error())
			}
		}

		timeoutSecs, err := cmd.Flags().GetInt("timeout")
		if err != nil {
			log.Fatalf("%v: %v", cmd.Use, err.Error())
		}

		if err := core.Instance().Processes().Start(ps, time.Duration(timeoutSecs)*time.Second); err != nil {
			log.Fatalf("%v: %v", cmd.Use, err.Error())
		}
	},
}

func init() {
	// flags
	startCmd.Flags().SortFlags = false

	startCmd.Flags().Int("timeout", 30, "timeout in seconds to wait for process to be up (0 = no timeout)\n")

	logger.AddFlags(startCmd.Flags())
}

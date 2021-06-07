package processes

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/osixia/container-baseimage/cmd/logger"
	"github.com/osixia/container-baseimage/config"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/log"
)

var startCmd = &cobra.Command{
	Use:   fmt.Sprintf("start process|%vname [process|%vname]...", config.TagsNamePrefix, config.TagsNamePrefix),
	Short: "Start process(es)",

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		ps, err := core.Instance().Processes().List(core.WithProcessesNames(args), core.HandleProcessesTagPrefixInNames(config.TagsNamePrefix))
		if err != nil {
			log.Fatalf("%v: %v", cmd.Use, err.Error())
		}

		// if no process defined get all processes
		if len(ps) == 0 {
			ps, err = core.Instance().Processes().List()
			if err != nil {
				log.Fatalf("%v: %v", cmd.Use, err.Error())
			}
		}

		if err := core.Instance().Processes().Start(ps); err != nil {
			log.Fatalf("%v: %v", cmd.Use, err.Error())
		}
	},
}

func init() {
	// flags
	startCmd.Flags().SortFlags = false
	logger.AddFlags(startCmd.Flags())
}

package processes

import (
	"fmt"

	"github.com/spf13/cobra"

	cmdlog "github.com/osixia/container-baseimage/cmd/log"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/helpers"
	"github.com/osixia/container-baseimage/log"
)

var statusCmd = &cobra.Command{
	Use:   fmt.Sprintf("status [process|%vname]...", TagNamePrefix),
	Short: "Print processes status",

	Long: "Print processes status\nWith no argument: print all processes status",

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		ps := helpers.MustVal(core.Instance().Processes().List(core.WithProcessesNames(args), core.HandleProcessesTagPrefixInNames(TagNamePrefix)))

		// if no process defined: get all processes
		if len(ps) == 0 {
			ps = helpers.MustVal(core.Instance().Processes().List())
		}

		if len(ps) == 0 {
			fmt.Println("No processes found")
			return
		}

		for _, p := range ps {
			fmt.Println(p.Status())
		}
	},
}

func init() {
	// flags
	statusCmd.Flags().SortFlags = false
	cmdlog.AddFlags(statusCmd.Flags())
}

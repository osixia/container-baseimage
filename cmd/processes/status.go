package processes

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	cmdlog "github.com/osixia/container-baseimage/cmd/log"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/helpers"
	"github.com/osixia/container-baseimage/log"
)

type statusFlags struct {
	json bool
}

var statusCmdFlags = &statusFlags{}

type processStatus struct {
	Name       string   `json:"name"`
	Up         bool     `json:"up"`
	Starting   bool     `json:"starting"`
	WantedDown bool     `json:"wantedDown"`
	Tags       []string `json:"tags"`
}

var statusCmd = &cobra.Command{
	Use:   fmt.Sprintf("status [process|%vname]...", TagNamePrefix),
	Short: "Print processes status",

	Long: "Print processes status\nWith no argument: print all processes status",

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		ps := helpers.MustVal(core.Instance().Processes().List(core.WithProcessesNames(args), core.HandleProcessesTagPrefixInNames(TagNamePrefix)))

		// if no arguments: get all processes
		if len(args) == 0 {
			ps = helpers.MustVal(core.Instance().Processes().List())
		}

		if statusCmdFlags.json {
			pss := make([]processStatus, 0, len(ps))
			for _, p := range ps {
				pss = append(pss, processStatus{Name: p.Name(), Up: p.IsUp(), Starting: p.IsStarting(), WantedDown: p.IsWantedDown(), Tags: p.Tags()})
			}

			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			helpers.Must(encoder.Encode(pss))
			return
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

	statusCmd.Flags().BoolVar(&statusCmdFlags.json, "json", false, "print status as JSON\n")

	cmdlog.AddFlags(statusCmd.Flags())
}

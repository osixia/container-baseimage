package processes

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	cmdlog "github.com/osixia/container-baseimage/cmd/log"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/helpers"
	"github.com/osixia/container-baseimage/log"
)

var checkCmd = &cobra.Command{
	Use:   fmt.Sprintf("check [process|%vname]...", TagNamePrefix),
	Short: "Check processes are up",

	Long: "Check processes are up, for container health checks\nExit with code 1 if a process is down, intentionally stopped processes excepted\nWith no argument: check processes of services linked to the entrypoint",

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		ps := helpers.MustVal(core.Instance().Processes().List(core.WithProcessesNames(args), core.HandleProcessesTagPrefixInNames(TagNamePrefix)))

		// if no arguments: get processes of linked services, the ones run by the
		// entrypoint by default. Listing runtime files would also check the
		// processes of other containers sharing the processes directory.
		if len(args) == 0 {
			ss := helpers.MustVal(core.Instance().Services().List(core.WithServicesLinked(true)))

			for _, s := range ss {
				if s.ProcessFile() == "" {
					continue
				}

				ps = append(ps, helpers.MustVal(core.Instance().Processes().Get(s.Name())))
			}
		} else if len(ps) == 0 {
			log.Fatalf("No processes found matching %s", strings.Join(args, ", "))
		}

		down := 0
		for _, p := range ps {
			if p.IsWantedDown() || p.IsUp() {
				continue
			}

			log.Errorf("%v: process is down", p.Name())
			down++
		}

		if down > 0 {
			os.Exit(1)
		}
	},
}

func init() {
	// flags
	checkCmd.Flags().SortFlags = false
	cmdlog.AddFlags(checkCmd.Flags())
}

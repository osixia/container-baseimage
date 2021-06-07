package services

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

type serviceStatus struct {
	Name       string   `json:"name"`
	Priority   int      `json:"priority"`
	Optional   bool     `json:"optional"`
	Downloaded bool     `json:"downloaded"`
	Installed  bool     `json:"installed"`
	Linked     bool     `json:"linked"`
	Tags       []string `json:"tags"`
}

var statusCmd = &cobra.Command{
	Use:   fmt.Sprintf("status [service|%vname]...", TagNamePrefix),
	Short: "Print services status",

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		ss := helpers.MustVal(core.Instance().Services().List(core.WithServicesNames(args), core.HandleServicesTagPrefixInNames(TagNamePrefix)))

		// if no arguments: get all services
		if len(args) == 0 {
			ss = helpers.MustVal(core.Instance().Services().List())
		}

		if statusCmdFlags.json {
			sss := make([]serviceStatus, 0, len(ss))
			for _, s := range ss {
				sss = append(sss, serviceStatus{Name: s.Name(), Priority: s.Priority(), Optional: s.IsOptional(), Downloaded: s.IsDownloaded(), Installed: s.IsInstalled(), Linked: s.IsLinked(), Tags: s.Tags()})
			}

			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			helpers.Must(encoder.Encode(sss))
			return
		}

		if len(ss) == 0 {
			fmt.Println("No services found")
			return
		}

		for _, s := range ss {
			fmt.Println(s.Status())
		}
	},
}

func init() {
	// flags
	statusCmd.Flags().SortFlags = false

	statusCmd.Flags().BoolVar(&statusCmdFlags.json, "json", false, "print status as JSON\n")

	cmdlog.AddFlags(statusCmd.Flags())
}

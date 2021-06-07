package entrypoint

import (
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/log"
)

var containerCmd = &cobra.Command{
	Use:   "container",
	Short: "Container image information",

	Aliases: []string{
		"c",
	},
}

var debugPackages = func() string {
	return strings.Join(core.Instance().Distribution().Config().DebugPackages, "\n")
}

var environmentFilesFunc = func() string {
	efs, err := core.Instance().Filesystem().ListDotEnv()
	if err != nil {
		log.Fatal(err.Error())
	}
	return strings.Join(efs, "\n")
}

var servicesFunc = func() string {
	svcs, err := core.Instance().Services().List()
	if err != nil {
		log.Fatal(err.Error())
	}

	r := ""
	for _, s := range svcs {
		mps := slices.Contains(core.Instance().Distribution().Config().MultiprocessStackServices, s.Name())
		els, _ := core.Instance().Entrypoint().ExistsLinkedService(s.Name())

		r += fmt.Sprintf("%v optional:%v,installed:%v,multiprocess-stack:%v,linked-to-entrypoint:%v\n", s.Name(), s.IsOptional(), s.IsInstalled(), mps, els)
	}

	return r
}

func init() {
	// subcommands
	containerCmd.AddCommand(newPrintCmd("debug-packages", "Debug packages", "dp", debugPackages))
	containerCmd.AddCommand(newPrintCmd("environment-files", "Environment file(s)", "ef", environmentFilesFunc))
	containerCmd.AddCommand(newPrintCmd("services", "Services", "s", servicesFunc))
}

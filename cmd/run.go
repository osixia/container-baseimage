package cmd

import (
	"os"
	"strings"

	"github.com/spf13/cobra"

	cmdlog "github.com/osixia/container-baseimage/cmd/log"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/helpers"
	"github.com/osixia/container-baseimage/log"
)

type runFlags struct {
	skipEnvFiles bool
}

var runCmdFlags = &runFlags{}

var runCmd = &cobra.Command{
	Use:   "run [args]...",
	Short: "Run a command instead of services",

	GroupID: mainGroupID,

	Aliases: []string{
		"c",
		"cmd",
	},

	Args: cobra.MinimumNArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		// set environment variables from environment files
		if !runCmdFlags.skipEnvFiles {
			files, err := core.Instance().Filesystem().ListDotEnv()
			if err != nil {
				log.Fatalf("%v: %v", cmd.Use, err.Error())
			}

			if err := core.Instance().Filesystem().LoadDotEnv(files); err != nil {
				log.Fatalf("%v: %v", cmd.Use, err.Error())
			}
		} else {
			log.Debug("Skipping getting environment variables values from environment files ...")
		}

		// log environment variables values
		log.Debugf("Environment variables:\n%v", strings.Join(os.Environ(), "\n"))

		if err := helpers.NewExec(cmd.Context()).WithNoLog(true).Command(args[0], args[1:]...).Run(); err != nil {
			log.Fatalf("%v: %v", cmd.Use, err.Error())
		}
	},
}

func init() {
	// flags
	runCmd.Flags().SortFlags = false

	runCmd.Flags().BoolVarP(&runCmdFlags.skipEnvFiles, "skip-env-files", "e", false, "skip loading environment variables from environment files\n")

	cmdlog.AddFlags(runCmd.Flags())
}

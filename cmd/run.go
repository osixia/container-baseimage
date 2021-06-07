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

	Args: cobra.MinimumNArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		// set environment variables from environment files
		if !runCmdFlags.skipEnvFiles {
			files := helpers.MustVal(core.Instance().Filesystem().ListDotEnv())

			log.Debugf("Loading environment variables from %v ...", strings.Join(files, ", "))
			helpers.Must(core.Instance().Filesystem().LoadDotEnv(files))

		} else {
			log.Debug("Skipping getting environment variables values from environment files ...")
		}

		// log environment variables values
		log.Debugf("Environment variables:\n%v", strings.Join(os.Environ(), "\n"))

		helpers.Must(helpers.NewExec(cmd.Context()).WithNoLog(true).Command(args[0], args[1:]...).Run())
	},
}

func init() {
	// flags
	runCmd.Flags().SortFlags = false

	runCmd.Flags().BoolVarP(&runCmdFlags.skipEnvFiles, "skip-env-files", "e", false, "skip loading environment variables from environment files\n")

	cmdlog.AddFlags(runCmd.Flags())
}

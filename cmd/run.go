package cmd

import (
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	cmdlog "github.com/osixia/container-baseimage/cmd/log"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/errors"
	"github.com/osixia/container-baseimage/helpers"
	"github.com/osixia/container-baseimage/log"
)

type runFlags struct {
	skipEnvFiles         bool
	killAllOnExitTimeout time.Duration
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

			if len(files) > 0 {
				log.Debugf("Loading environment variables from %v ...", strings.Join(files, ", "))
				helpers.Must(core.Instance().Filesystem().LoadDotEnv(files))
			}

		} else {
			log.Debug("Skipping getting environment variables values from environment files ...")
		}

		// log environment variables values
		log.Debugf("Environment variables:\n%v", strings.Join(os.Environ(), "\n"))

		ctx, stop := signal.NotifyContext(cmd.Context(), core.LifecycleInterceptedSignals...)

		// Go exits on some signals (e.g. SIGHUP): catch them all before
		// starting the command, then forward them like the entrypoint does.
		sigChan := make(chan os.Signal, 32)
		signal.Notify(sigChan)

		e := helpers.NewExec(ctx).WithSetGPID(!helpers.IsTerminal(os.Stdin)).WithStopSignal(syscall.SIGTERM).WithTimeout(runCmdFlags.killAllOnExitTimeout).WithNoLog(true).Command(args[0], args[1:]...)

		err := e.Start()
		if err == nil {
			go core.ForwardSignals(sigChan, e)

			err = e.Wait()
		}
		signal.Stop(sigChan)
		close(sigChan)
		stop()
		if err != nil {
			log.Error(err.Error())
		}
		os.Exit(errors.ExitCode(err))
	},
}

func init() {
	// flags
	runCmd.Flags().SortFlags = false

	runCmd.Flags().BoolVarP(&runCmdFlags.skipEnvFiles, "skip-env-files", "e", false, "skip loading environment variables from environment files\n")
	runCmd.Flags().DurationVarP(&runCmdFlags.killAllOnExitTimeout, "kill-all-timeout", "t", 15*time.Second, "force kill command after timeout (SIGKILL after delay)\n")

	cmdlog.AddFlags(runCmd.Flags())
}

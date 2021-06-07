package entrypoint

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/osixia/container-baseimage/cmd/entrypoint/generate"
	"github.com/osixia/container-baseimage/cmd/logger"
	"github.com/osixia/container-baseimage/config"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/log"
)

const (
	Banner = " / _ \\ ___(_)_  _(_) __ _   / / __ )  __ _ ___  ___(_)_ __ ___   __ _  __ _  ___ \n| | | / __| \\ \\/ / |/ _` | / /|  _ \\ / _` / __|/ _ \\ | '_ ` _ \\ / _` |/ _` |/ _ \n| |_| \\__ \\ |>  <| | (_| |/ / | |_) | (_| \\__ \\  __/ | | | | | | (_| | (_| |  __/\n \\___/|___/_/_/\\_\\_|\\__,_/_/  |____/ \\__,_|___/\\___|_|_| |_| |_|\\__,_|\\__, |\\___|\n                                                                      |___/      "
)

type entrypointFlags struct {
	core.EntrypointOptions

	runOnlyLifecycleStep string
	debug                bool
	version              bool
}

var ErrInvalidLifecycleStep = errors.New("invalid lifecycle step")

var runOnlyLifecycleSteps = []string{
	core.LifecycleStepStartup.String(),
	core.LifecycleStepProcess.String(),
	core.LifecycleStepFinish.String(),
}

var entrypointCmdFlags = &entrypointFlags{}

var EntrypointCmd = &cobra.Command{
	Use:   "entrypoint",
	Short: "Container entrypoint",

	Long: fmt.Sprintf("\n%v\nContainer image built with osixia/baseimage (%v) 🐳✨🌴\nhttps://github.com/osixia/container-baseimage", Banner, config.BuildVersion),

	Aliases: []string{
		"e",
	},

	PreRunE: func(cmd *cobra.Command, args []string) error {
		log.Tracef("PreRunE: %v called with args: %v", cmd.Use, args)

		return entrypointCmdFlags.validate(core.Instance().Services())
	},

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		containerImage := core.Instance().Config().Image

		if entrypointCmdFlags.version {
			fmt.Println(containerImage)
			os.Exit(0)
		}
		log.Infof("Container image: %v", containerImage)

		if entrypointCmdFlags.debug && log.Level() < log.LevelDebug {
			if err := log.SetLevel(log.Levels[log.LevelDebug]); err != nil {
				log.Error(err.Error())
			}
		}

		epo := entrypointCmdFlags.toEntrypointOptions()
		epo.Commands = args

		exitCode, err := core.Instance().Entrypoint().Run(cmd.Context(), epo)
		if err != nil {
			log.Fatalf("%v: %v", cmd.Use, err.Error())
		}

		os.Exit(exitCode)
	},
}

func init() {
	// subcommands
	EntrypointCmd.AddCommand(linkToEntrypointCmd)
	EntrypointCmd.AddCommand(unlinkToEntrypointCmd)
	EntrypointCmd.AddCommand(generate.GenerateCmd)
	EntrypointCmd.AddCommand(containerCmd)
	EntrypointCmd.AddCommand(thanksCmd)

	// flags
	EntrypointCmd.Flags().SortFlags = false

	EntrypointCmd.Flags().BoolVarP(&entrypointCmdFlags.SkipEnvFiles, "skip-env-files", "e", false, "skip getting environment variables values from environment file(s)")

	EntrypointCmd.Flags().BoolVarP(&entrypointCmdFlags.SkipStartup, "skip-startup", "s", false, "skip running pre-startup-cmd and service(s) startup.sh script(s)")
	EntrypointCmd.Flags().BoolVarP(&entrypointCmdFlags.SkipProcess, "skip-process", "p", false, "skip running pre-process-cmd and service(s) process.sh script(s)")
	EntrypointCmd.Flags().BoolVarP(&entrypointCmdFlags.SkipFinish, "skip-finish", "f", false, "skip running pre-finish-cmd and service(s) finish.sh script(s)")
	EntrypointCmd.Flags().StringVarP(&entrypointCmdFlags.runOnlyLifecycleStep, "run-only-lifecycle-step", "c", "", fmt.Sprintf("run only one lifecycle step pre-command and script(s) file(s), choices: %v\n", strings.Join(runOnlyLifecycleSteps, ", ")))
	EntrypointCmd.MarkFlagsMutuallyExclusive("run-only-lifecycle-step", "skip-startup")
	EntrypointCmd.MarkFlagsMutuallyExclusive("run-only-lifecycle-step", "skip-process")
	EntrypointCmd.MarkFlagsMutuallyExclusive("run-only-lifecycle-step", "skip-finish")

	EntrypointCmd.Flags().StringArrayVarP(&entrypointCmdFlags.PreStartupCmds, "pre-startup-cmd", "1", nil, "run command passed as argument before service(s) startup.sh script(s)")
	EntrypointCmd.Flags().StringArrayVarP(&entrypointCmdFlags.PreProcessCmds, "pre-process-cmd", "3", nil, "run command passed as argument before service(s) process.sh script(s)")
	EntrypointCmd.Flags().StringArrayVarP(&entrypointCmdFlags.PreFinishCmds, "pre-finish-cmd", "5", nil, "run command passed as argument before service(s) finish.sh script(s)")
	EntrypointCmd.Flags().StringArrayVarP(&entrypointCmdFlags.PreExitCmds, "pre-exit-cmd", "7", nil, "run command passed as argument before container exits\n")

	EntrypointCmd.Flags().StringArrayVarP(&entrypointCmdFlags.Services, "service", "x", nil, "run only listed service(s), (default: run all container services)\n")

	EntrypointCmd.Flags().BoolVarP(&entrypointCmdFlags.RunBash, "bash", "b", false, "run bash in addition to service(s) or command\n")

	EntrypointCmd.Flags().BoolVarP(&entrypointCmdFlags.TerminateAllOnExit, "kill-all-on-exit", "k", true, "kill all processes on the system upon exiting (send sigterm to all processes)")
	EntrypointCmd.Flags().DurationVarP(&entrypointCmdFlags.TerminateAllOnExitTimeout, "kill-all-on-exit-timeout", "t", 15*time.Second, "kill all processes timeout (send sigkill to all processes after sigterm timeout has been reached)")
	EntrypointCmd.Flags().BoolVarP(&entrypointCmdFlags.RestartProcesses, "restart-processes", "r", true, "automatically restart failed services process.sh scripts (multiprocess container image only)")
	EntrypointCmd.Flags().BoolVarP(&entrypointCmdFlags.KeepAlive, "keep-alive", "a", false, "keep alive container after all processes have exited\n")

	EntrypointCmd.Flags().BoolVarP(&entrypointCmdFlags.UnsecureFastWrite, "unsecure-fast-write", "w", false, "disable fsync and friends with eatmydata LD_PRELOAD library\n")

	EntrypointCmd.Flags().BoolVarP(&entrypointCmdFlags.debug, "debug", "d", false, "set log level to debug and install debug packages")
	EntrypointCmd.Flags().StringSliceVarP(&entrypointCmdFlags.InstallPackages, "install-packages", "i", nil, "install packages\n")

	EntrypointCmd.Flags().BoolVarP(&entrypointCmdFlags.version, "version", "v", false, "print container image version\n")

	logger.AddFlags(EntrypointCmd.Flags())
}

func (o *entrypointFlags) validate(s core.Services) error {

	if o.runOnlyLifecycleStep != "" {
		if err := o.validateRunOnlyLifecycleStep(); err != nil {
			return err
		}
	}

	return o.EntrypointOptions.Validate(s)
}

func (o *entrypointFlags) validateRunOnlyLifecycleStep() error {

	for _, v := range runOnlyLifecycleSteps {
		if v == o.runOnlyLifecycleStep {
			return nil
		}
	}

	return fmt.Errorf("%v: %w (choices: %v)", o.runOnlyLifecycleStep, ErrInvalidLifecycleStep, strings.Join(runOnlyLifecycleSteps, ", "))
}

func (o *entrypointFlags) toEntrypointOptions() core.EntrypointOptions {

	if o.debug {
		// append debug packages to packages to install
		o.InstallPackages = append(o.InstallPackages, core.Instance().Distribution().Config().DebugPackages...)
	}

	switch o.runOnlyLifecycleStep {
	case "startup":
		o.SkipProcess = true
		o.SkipFinish = true
	case "process":
		o.SkipStartup = true
		o.SkipFinish = true
	case "finish":
		o.SkipStartup = true
		o.SkipProcess = true
	}

	return o.EntrypointOptions
}

func newPrintCmd(use string, short string, alias string, f func() string) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,

		Aliases: []string{
			alias,
		},

		Run: func(cmd *cobra.Command, args []string) {
			log.Tracef("Run: %v called with args: %v", cmd.Use, args)

			fmt.Println(f())
		},
	}
}

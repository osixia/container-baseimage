package entrypoint

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/osixia/container-baseimage/cmd/entrypoint/generate"
	"github.com/osixia/container-baseimage/cmd/logger"
	"github.com/osixia/container-baseimage/cmd/services"
	"github.com/osixia/container-baseimage/config"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/log"
)

const (
	Banner = " / _ \\ ___(_)_  _(_) __ _   / / __ )  __ _ ___  ___(_)_ __ ___   __ _  __ _  ___ \n| | | / __| \\ \\/ / |/ _` | / /|  _ \\ / _` / __|/ _ \\ | '_ ` _ \\ / _` |/ _` |/ _ \n| |_| \\__ \\ |>  <| | (_| |/ / | |_) | (_| \\__ \\  __/ | | | | | | (_| | (_| |  __/\n \\___/|___/_/_/\\_\\_|\\__,_/_/  |____/ \\__,_|___/\\___|_|_| |_| |_|\\__,_|\\__, |\\___|\n                                                                      |___/      "
)

var ErrInvalidLifecycleStep = errors.New("invalid lifecycle step")

var entrypointCmdFlags = &entrypointFlags{}

type NilBool struct {
	Value *bool
}

func (b *NilBool) Set(s string) error {

	if s == "" {
		s = "true"
	}

	v, err := strconv.ParseBool(s)
	if err != nil {
		return err
	}
	b.Value = &v

	return nil
}

func (b *NilBool) String() string {
	if b.Value == nil {
		return ""
	}
	return fmt.Sprintf("%t", *b.Value)
}

func (b *NilBool) Type() string {
	return "bool"
}

type entrypointFlags struct {
	core.EntrypointOptions

	exec                 []string
	runOnlyLifecycleStep string
	restart              NilBool
	debug                bool
	version              bool
}

func (o *entrypointFlags) toEntrypointOptions() (core.EntrypointOptions, error) {

	var svcs []core.Service
	var err error

	if o.exec != nil {
		log.Trace("search services to exec")

		svcs, err = core.Instance().Services().List(core.WithServicesNames(o.exec), core.HandleServicesTagPrefixInNames(services.TagNamePrefix))
		if err != nil {
			return o.EntrypointOptions, err
		}
	} else {
		log.Trace("no service is specified, search all services linked to the entrypoint")

		svcs, err = core.Instance().Services().List(core.WithServicesLinked(true))
		if err != nil {
			return o.EntrypointOptions, err
		}
	}

	log.Tracef("services found: %v", svcs)
	o.Services = svcs

	if o.runOnlyLifecycleStep != "" {
		switch o.runOnlyLifecycleStep {
		case string(core.LifecycleStepStartup):
			o.SkipProcess = true
			o.SkipFinish = true
		case string(core.LifecycleStepProcess):
			o.SkipStartup = true
			o.SkipFinish = true
		case string(core.LifecycleStepFinish):
			o.SkipStartup = true
			o.SkipProcess = true
		default:
			return o.EntrypointOptions, fmt.Errorf("%v: %w (choices: %v)", o.runOnlyLifecycleStep, ErrInvalidLifecycleStep, strings.Join(core.LifecycleStepsList(), ", "))
		}
	}

	o.RestartProcesses = o.restart.Value

	if o.debug {
		// force log level to debug at least
		if log.Level() < log.LevelDebug {
			if err := log.SetLevel(log.Levels[log.LevelDebug]); err != nil {
				log.Error(err.Error())
			}
		}

		// append debug packages to packages to install
		o.InstallPackages = append(o.InstallPackages, core.Instance().Distribution().Config().DebugPackages...)

		// run bash
		o.RunBash = true
	}

	// skip process is set but we want to run bash
	if o.SkipProcess && o.RunBash {
		// empty services to run
		o.Services = nil

		// set skip process to false so bash is run
		o.SkipProcess = false
	}

	return o.EntrypointOptions, nil
}

var EntrypointCmd = &cobra.Command{
	Use:   "entrypoint",
	Short: "Container entrypoint",

	Long: fmt.Sprintf("\n%v\nContainer image built with osixia/baseimage (%v) 🐳✨🌴\nhttps://github.com/osixia/container-baseimage", Banner, config.Version),

	Aliases: []string{
		"ep",
	},

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		containerImage := core.Instance().Config().Image

		if entrypointCmdFlags.version {
			fmt.Println(containerImage)
			os.Exit(0)
		}
		log.Infof("Container image: %v", containerImage)

		epo, err := entrypointCmdFlags.toEntrypointOptions()
		if err != nil {
			log.Fatalf("%v: %v", cmd.Use, err.Error())
		}

		epo.Args = args

		exitCode, err := core.Instance().Entrypoint().Run(cmd.Context(), epo)
		if err != nil {
			log.Error(err.Error())
		}

		os.Exit(exitCode)
	},
}

func init() {
	// subcommands
	EntrypointCmd.AddCommand(generate.GenerateCmd)
	EntrypointCmd.AddCommand(thanksCmd)

	// flags
	EntrypointCmd.Flags().SortFlags = false

	EntrypointCmd.Flags().BoolVarP(&entrypointCmdFlags.SkipEnvFiles, "skip-env-files", "e", false, "skip getting environment variables values from environment file(s)\n")

	EntrypointCmd.Flags().BoolVarP(&entrypointCmdFlags.SkipStartup, "skip-startup", "s", false, "skip running pre-startup-cmd and service(s) startup.sh script(s)")
	EntrypointCmd.Flags().BoolVarP(&entrypointCmdFlags.SkipProcess, "skip-process", "p", false, "skip running pre-process-cmd and service(s) process.sh script(s)")
	EntrypointCmd.Flags().BoolVarP(&entrypointCmdFlags.SkipFinish, "skip-finish", "f", false, "skip running pre-finish-cmd and service(s) finish.sh script(s)")
	EntrypointCmd.Flags().StringVarP(&entrypointCmdFlags.runOnlyLifecycleStep, "run-only-lifecycle-step", "c", "", fmt.Sprintf("run only one lifecycle step pre-command and script(s) file(s), choices: %v\n", strings.Join(core.LifecycleStepsList(), ", ")))
	EntrypointCmd.MarkFlagsMutuallyExclusive("run-only-lifecycle-step", "skip-startup")
	EntrypointCmd.MarkFlagsMutuallyExclusive("run-only-lifecycle-step", "skip-process")
	EntrypointCmd.MarkFlagsMutuallyExclusive("run-only-lifecycle-step", "skip-finish")

	EntrypointCmd.Flags().StringArrayVarP(&entrypointCmdFlags.PreStartupCmds, "pre-startup-cmd", "1", nil, "run command passed as argument before service(s) startup.sh script(s)")
	EntrypointCmd.Flags().StringArrayVarP(&entrypointCmdFlags.PreProcessCmds, "pre-process-cmd", "3", nil, "run command passed as argument before service(s) process.sh script(s)")
	EntrypointCmd.Flags().StringArrayVarP(&entrypointCmdFlags.PreFinishCmds, "pre-finish-cmd", "5", nil, "run command passed as argument before service(s) finish.sh script(s)")
	EntrypointCmd.Flags().StringArrayVarP(&entrypointCmdFlags.PreExitCmds, "pre-exit-cmd", "7", nil, "run command passed as argument before container exits\n")

	EntrypointCmd.Flags().StringArrayVarP(&entrypointCmdFlags.exec, "exec", "x", nil, "execute only listed service(s) (default run service(s) linked to the entrypoint)\n")

	EntrypointCmd.Flags().BoolVarP(&entrypointCmdFlags.RunBash, "bash", "b", false, "run Bash along with other service(s) or command\n")

	EntrypointCmd.Flags().BoolVarP(&entrypointCmdFlags.TerminateAllOnExit, "kill-all-on-exit", "k", true, "kill all processes on the system upon exiting (send sigterm to all processes)")
	EntrypointCmd.Flags().DurationVarP(&entrypointCmdFlags.TerminateAllOnExitTimeout, "kill-all-on-exit-timeout", "t", 15*time.Second, "kill all processes timeout (send sigkill to all processes after sigterm timeout has been reached)")
	EntrypointCmd.Flags().VarP(&entrypointCmdFlags.restart, "restart", "r", "automatically restart failed services process.sh scripts (single-process: default false, multi-process: default true)")
	EntrypointCmd.Flags().BoolVarP(&entrypointCmdFlags.KeepAlive, "keep-alive", "a", false, "keep alive container after all processes have exited\n")

	EntrypointCmd.Flags().BoolVarP(&entrypointCmdFlags.UnsecureFastWrite, "unsecure-fast-write", "w", false, "disable fsync and friends with eatmydata LD_PRELOAD library\n")

	EntrypointCmd.Flags().BoolVarP(&entrypointCmdFlags.debug, "debug", "d", false, "set log level to debug, install debug packages and run Bash")
	EntrypointCmd.Flags().StringSliceVarP(&entrypointCmdFlags.InstallPackages, "install-packages", "i", nil, "install packages\n")

	EntrypointCmd.Flags().BoolVarP(&entrypointCmdFlags.version, "version", "v", false, "print container image version\n")

	logger.AddFlags(EntrypointCmd.Flags())
}

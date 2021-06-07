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

	step string

	exec    []string
	skip    []string
	skipAll bool

	restart NilBool
	debug   bool
	version bool
}

func (o *entrypointFlags) setServices() error {

	if o.skipAll {
		log.Trace("skip all services")
		return nil
	}

	var err error

	if len(o.exec) == 0 && len(o.skip) == 0 {
		log.Trace("no service is specified, to exec or skip search all services linked to the entrypoint")

		o.Services, err = core.Instance().Services().List(core.WithServicesLinked(true))
		if err != nil {
			return err
		}

		return nil
	}

	if len(o.exec) > 0 {
		log.Trace("search services to exec")

		o.Services, err = core.Instance().Services().List(core.WithServicesNames(o.exec), core.HandleServicesTagPrefixInNames(services.TagNamePrefix))
		if err != nil {
			return err
		}

		return nil
	}

	if len(o.skip) > 0 {
		o.Services, err = core.Instance().Services().List(core.WithoutServicesNames(o.skip), core.HandleServicesTagPrefixInNames(services.TagNamePrefix))
		if err != nil {
			return err
		}

		return nil
	}

	return nil
}

func (o *entrypointFlags) toEntrypointOptions() (core.EntrypointOptions, error) {

	if err := o.setServices(); err != nil {
		return o.EntrypointOptions, err
	}

	if o.step != "" {
		switch o.step {
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
			return o.EntrypointOptions, fmt.Errorf("%v: %w (choices: %v)", o.step, ErrInvalidLifecycleStep, strings.Join(core.LifecycleStepsList(), ", "))
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

	// flags
	EntrypointCmd.Flags().SortFlags = false

	EntrypointCmd.Flags().BoolVarP(&entrypointCmdFlags.SkipEnvFiles, "skip-env-files", "e", false, "skip loading environment variables from environment files\n")

	EntrypointCmd.Flags().BoolVarP(&entrypointCmdFlags.SkipStartup, "skip-startup", "s", false, "skip running pre-startup-cmd and startup scripts")
	EntrypointCmd.Flags().BoolVarP(&entrypointCmdFlags.SkipProcess, "skip-process", "p", false, "skip running pre-process-cmd and process scripts")
	EntrypointCmd.Flags().BoolVarP(&entrypointCmdFlags.SkipFinish, "skip-finish", "f", false, "skip running pre-finish-cmd and finish scripts")
	EntrypointCmd.Flags().StringVar(&entrypointCmdFlags.step, "step", "", fmt.Sprintf("run only one lifecycle step with its pre-commands and scripts, choices: %v\n", strings.Join(core.LifecycleStepsList(), ", ")))
	EntrypointCmd.MarkFlagsMutuallyExclusive("step", "skip-startup")
	EntrypointCmd.MarkFlagsMutuallyExclusive("step", "skip-process")
	EntrypointCmd.MarkFlagsMutuallyExclusive("step", "skip-finish")

	EntrypointCmd.Flags().StringArrayVar(&entrypointCmdFlags.PreStartupCmds, "pre-startup-cmd", nil, "run command before startup scripts")
	EntrypointCmd.Flags().StringArrayVar(&entrypointCmdFlags.PreProcessCmds, "pre-process-cmd", nil, "run command before process scripts")
	EntrypointCmd.Flags().StringArrayVar(&entrypointCmdFlags.PreFinishCmds, "pre-finish-cmd", nil, "run command before finish scripts")
	EntrypointCmd.Flags().StringArrayVar(&entrypointCmdFlags.PreExitCmds, "pre-exit-cmd", nil, "run command before container exits\n")

	EntrypointCmd.Flags().StringArrayVarP(&entrypointCmdFlags.exec, "exec", "x", nil, "execute only listed services (default: run services linked to the entrypoint)")
	EntrypointCmd.Flags().StringArrayVarP(&entrypointCmdFlags.skip, "skip", "X", nil, "skip listed services")
	EntrypointCmd.Flags().BoolVarP(&entrypointCmdFlags.skipAll, "skip-all", "", false, "skip all services\n")
	EntrypointCmd.MarkFlagsMutuallyExclusive("exec", "skip")
	EntrypointCmd.MarkFlagsMutuallyExclusive("exec", "skip-all")
	EntrypointCmd.MarkFlagsMutuallyExclusive("skip", "skip-all")

	EntrypointCmd.Flags().BoolVarP(&entrypointCmdFlags.RunBash, "bash", "b", false, "run Bash alongside other services or command\n")

	EntrypointCmd.Flags().BoolVarP(&entrypointCmdFlags.TerminateAllOnExit, "kill-all-on-exit", "k", true, "kill all remaining container processes on exit (send SIGTERM)")
	EntrypointCmd.Flags().DurationVarP(&entrypointCmdFlags.TerminateAllOnExitTimeout, "kill-all-on-exit-timeout", "t", 15*time.Second, "force kill all processes after timeout (send SIGKILL after SIGTERM timeout)")
	EntrypointCmd.Flags().VarP(&entrypointCmdFlags.restart, "restart", "r", "automatically restart failed processes (single-process: default false, multi-process: default true)")
	EntrypointCmd.Flags().BoolVarP(&entrypointCmdFlags.KeepAlive, "keep-alive", "a", false, "keep container alive after all processes have exited\n")

	EntrypointCmd.Flags().BoolVarP(&entrypointCmdFlags.UnsecureFastWrite, "unsafe-fast-write", "w", false, "disable fsync and related sync operations using eatmydata\n")

	EntrypointCmd.Flags().BoolVarP(&entrypointCmdFlags.debug, "debug", "d", false, "set log level to debug, install debug packages, and run Bash")
	EntrypointCmd.Flags().StringSliceVarP(&entrypointCmdFlags.InstallPackages, "install-packages", "i", nil, "install packages\n")

	EntrypointCmd.Flags().BoolVarP(&entrypointCmdFlags.version, "version", "v", false, "print container image version\n")

	logger.AddFlags(EntrypointCmd.Flags())
}

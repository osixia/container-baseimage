package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/osixia/container-baseimage/cmd/environment"
	"github.com/osixia/container-baseimage/cmd/envsubst"
	"github.com/osixia/container-baseimage/cmd/generate"
	"github.com/osixia/container-baseimage/cmd/groups"
	cmdlog "github.com/osixia/container-baseimage/cmd/log"
	"github.com/osixia/container-baseimage/cmd/packages"
	"github.com/osixia/container-baseimage/cmd/processes"
	"github.com/osixia/container-baseimage/cmd/services"
	"github.com/osixia/container-baseimage/cmd/users"
	"github.com/osixia/container-baseimage/config"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/helpers"
	"github.com/osixia/container-baseimage/log"
)

const (
	mainGroupID       = "setup"
	runtimeGroupID    = "runtime"
	envGroupID        = "env"
	filesystemGroupID = "fs"
	logGroupID        = "log"
)

const (
	Banner = " / _ \\ ___(_)_  _(_) __ _   / / __ )  __ _ ___  ___(_)_ __ ___   __ _  __ _  ___ \n| | | / __| \\ \\/ / |/ _` | / /|  _ \\ / _` / __|/ _ \\ | '_ ` _ \\ / _` |/ _` |/ _ \n| |_| \\__ \\ |>  <| | (_| |/ / | |_) | (_| \\__ \\  __/ | | | | | | (_| | (_| |  __/\n \\___/|___/_/_/\\_\\_|\\__,_/_/  |____/ \\__,_|___/\\___|_|_| |_| |_|\\__,_|\\__, |\\___|\n                                                                      |___/      "
)

var ErrInvalidLifecycleStep = errors.New("invalid lifecycle step")

var cmdFlags = &flags{}

type flags struct {
	core.EntrypointOptions

	step string

	exec    []string
	skip    []string
	skipAll bool

	restart helpers.NilBool

	version bool
}

func (o *flags) setServices() error {

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

func (o *flags) toEntrypointOptions() (*core.EntrypointOptions, error) {

	if err := o.setServices(); err != nil {
		return nil, err
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
			return nil, fmt.Errorf("%v: %w (choices: %v)", o.step, ErrInvalidLifecycleStep, strings.Join(core.LifecycleStepsList(), ", "))
		}
	}

	o.RestartProcesses = o.restart.Value

	// skip process is set but we want to run bash
	if o.SkipProcess && o.RunBash {
		// empty services to run
		o.Services = nil

		// set skip process to false so bash is run
		o.SkipProcess = false
	}

	return &o.EntrypointOptions, nil
}

var cmd = &cobra.Command{
	Use: "container",

	Long: fmt.Sprintf("\n%v\nBuilt with osixia/baseimage (%v) 🐳✨🌴\nhttps://github.com/osixia/container-baseimage\n\nContainer runtime with lifecycle management and system utilities", Banner, config.Version),

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return cmdlog.HandleFlags(cmd)
	},

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		containerImage := core.Instance().Config().Image

		if cmdFlags.version {
			fmt.Println(containerImage)
			os.Exit(0)
		}
		log.Infof("Container image: %v", containerImage)

		epo, err := cmdFlags.toEntrypointOptions()
		if err != nil {
			log.Fatalf("%v: %v", cmd.Use, err.Error())
		}

		epo.Args = args

		exitCode, err := core.Instance().Entrypoint().Run(cmd.Context(), *epo)
		if err != nil {
			log.Error(err.Error())
		}

		os.Exit(exitCode)
	},
}

func init() {
	cobra.EnableCommandSorting = false

	// subcommands groups
	cmd.AddGroup(&cobra.Group{
		ID:    mainGroupID,
		Title: "Main Commands:",
	})

	cmd.AddGroup(&cobra.Group{
		ID:    runtimeGroupID,
		Title: "Runtime Commands:",
	})

	cmd.AddGroup(&cobra.Group{
		ID:    envGroupID,
		Title: "Environment Commands:",
	})

	cmd.AddGroup(&cobra.Group{
		ID:    filesystemGroupID,
		Title: "Filesystem Commands:",
	})

	cmd.AddGroup(&cobra.Group{
		ID:    logGroupID,
		Title: "Logging Command:",
	})

	// subcommands
	generate.GenerateCmd.GroupID = mainGroupID
	cmd.AddCommand(installCmd)
	cmd.AddCommand(generate.GenerateCmd)
	cmd.AddCommand(runCmd)

	services.ServicesCmd.GroupID = runtimeGroupID
	processes.ProcessesCmd.GroupID = runtimeGroupID
	packages.PackagesCmd.GroupID = runtimeGroupID
	cmd.AddCommand(packages.PackagesCmd)
	cmd.AddCommand(processes.ProcessesCmd)
	cmd.AddCommand(services.ServicesCmd)

	environment.EnvironmentCmd.GroupID = envGroupID
	envsubst.EnvsubstCmd.GroupID = envGroupID
	cmd.AddCommand(environment.EnvironmentCmd)
	cmd.AddCommand(envsubst.EnvsubstCmd)

	groups.GroupsCmd.GroupID = filesystemGroupID
	users.UsersCmd.GroupID = filesystemGroupID
	cmd.AddCommand(groups.GroupsCmd)
	cmd.AddCommand(users.UsersCmd)
	cmd.AddCommand(watchCmd)

	cmdlog.LogCmd.GroupID = logGroupID
	cmd.AddCommand(cmdlog.LogCmd)

	// flags
	cmd.Flags().SortFlags = false

	cmd.Flags().BoolVarP(&cmdFlags.SkipEnvFiles, "skip-env-files", "e", false, "skip loading environment variables from environment files\n")

	cmd.Flags().BoolVarP(&cmdFlags.SkipStartup, "skip-startup", "s", false, "skip running pre-startup-cmd and startup scripts")
	cmd.Flags().BoolVarP(&cmdFlags.SkipProcess, "skip-process", "p", false, "skip running pre-process-cmd and process scripts")
	cmd.Flags().BoolVarP(&cmdFlags.SkipFinish, "skip-finish", "f", false, "skip running pre-finish-cmd and finish scripts")
	cmd.Flags().StringVar(&cmdFlags.step, "step", "", fmt.Sprintf("run a single lifecycle step, choices: %v\n", strings.Join(core.LifecycleStepsList(), ", ")))
	cmd.MarkFlagsMutuallyExclusive("step", "skip-startup")
	cmd.MarkFlagsMutuallyExclusive("step", "skip-process")
	cmd.MarkFlagsMutuallyExclusive("step", "skip-finish")

	cmd.Flags().StringArrayVar(&cmdFlags.PreStartupCmds, "pre-startup-cmd", nil, "run command before startup scripts")
	cmd.Flags().StringArrayVar(&cmdFlags.PreProcessCmds, "pre-process-cmd", nil, "run command before process scripts")
	cmd.Flags().StringArrayVar(&cmdFlags.PreFinishCmds, "pre-finish-cmd", nil, "run command before finish scripts")
	cmd.Flags().StringArrayVar(&cmdFlags.PreExitCmds, "pre-exit-cmd", nil, "run command before container exits\n")

	cmd.Flags().StringArrayVarP(&cmdFlags.exec, "exec", "x", nil, "execute only specified services (default services linked to the entrypoint)")
	cmd.Flags().StringArrayVarP(&cmdFlags.skip, "skip", "X", nil, "skip listed services")
	cmd.Flags().BoolVarP(&cmdFlags.skipAll, "skip-all", "", false, "skip all services\n")
	cmd.MarkFlagsMutuallyExclusive("exec", "skip")
	cmd.MarkFlagsMutuallyExclusive("exec", "skip-all")

	cmd.Flags().BoolVarP(&cmdFlags.RunBash, "bash", "b", false, "run Bash alongside other services or commands\n")

	cmd.Flags().BoolVarP(&cmdFlags.KillAllOnExit, "kill-all", "k", true, "kill all remaining container processes on exit (SIGTERM)")
	cmd.Flags().DurationVarP(&cmdFlags.KillAllOnExitTimeout, "kill-all-timeout", "t", 15*time.Second, "force kill remaining processes after timeout (SIGKILL after delay)")
	cmd.Flags().VarP(&cmdFlags.restart, "restart", "r", "restart failed processes (default single-process false, multi-process true)")
	cmd.Flags().BoolVarP(&cmdFlags.KeepAlive, "keep-alive", "a", false, "keep container alive after all processes have exited\n")

	cmd.Flags().BoolVarP(&cmdFlags.FastWrite, "fast-write", "w", false, "faster writes by disabling fsync (useful for CI, risk of data loss)\n")

	cmd.Flags().BoolVarP(&cmdFlags.version, "version", "v", false, "print container image version\n")

	cmdlog.AddFlags(cmd.Flags())
}

func Run(ctx context.Context) error {
	return cmd.ExecuteContext(ctx)
}

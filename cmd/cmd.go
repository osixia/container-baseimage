package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/osixia/container-baseimage/cmd/entrypoint"
	"github.com/osixia/container-baseimage/cmd/environment"
	"github.com/osixia/container-baseimage/cmd/envsubst"
	"github.com/osixia/container-baseimage/cmd/groups"
	cmdlog "github.com/osixia/container-baseimage/cmd/log"
	"github.com/osixia/container-baseimage/cmd/packages"
	"github.com/osixia/container-baseimage/cmd/processes"
	"github.com/osixia/container-baseimage/cmd/services"
	"github.com/osixia/container-baseimage/cmd/users"
)

const (
	setupGroupID      = "setup"
	entrypointGroupID = "entrypoint"
	runtimeGroupID    = "runtime"
	envGroupID        = "env"
	filesystemGroupID = "fs"
	logGroupID        = "log"
)

var cmd = &cobra.Command{
	Use: "container",

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return cmdlog.HandleFlags(cmd)
	},
}

func init() {
	cobra.EnableCommandSorting = false

	// subcommands groups
	cmd.AddGroup(&cobra.Group{
		ID:    setupGroupID,
		Title: "Setup Commands:",
	})

	cmd.AddGroup(&cobra.Group{
		ID:    entrypointGroupID,
		Title: "Entrypoint Command:",
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
		Title: "Log Command:",
	})

	// subcommands
	cmd.AddCommand(installCmd)
	cmd.AddCommand(debugCmd)

	entrypoint.EntrypointCmd.GroupID = entrypointGroupID
	cmd.AddCommand(entrypoint.EntrypointCmd)

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
}

func Run(ctx context.Context) error {
	return cmd.ExecuteContext(ctx)
}

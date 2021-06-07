package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/osixia/container-baseimage/cmd/entrypoint"
	"github.com/osixia/container-baseimage/cmd/logger"
	"github.com/osixia/container-baseimage/cmd/packages"
	"github.com/osixia/container-baseimage/cmd/processes"
	"github.com/osixia/container-baseimage/cmd/services"
)

const (
	installGroupID    = "install"
	entrypointGroupID = "entrypoint"
	runtimeGroupID    = "runtime"
	envGroupID        = "env"
	loggerGroupID     = "logger"
)

var cmd = &cobra.Command{
	Use: "container",

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return logger.HandleFlags(cmd)
	},
}

func init() {
	cobra.EnableCommandSorting = false

	// subcommands groups
	cmd.AddGroup(&cobra.Group{
		ID:    installGroupID,
		Title: "Install Commands:",
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
		ID:    loggerGroupID,
		Title: "Logger Command:",
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

	cmd.AddCommand(environmentFilesCmd)
	cmd.AddCommand(envsubstCmd)
	cmd.AddCommand(envsubstTemplatesCmd)

	logger.LoggerCmd.GroupID = loggerGroupID
	cmd.AddCommand(logger.LoggerCmd)
}

func Run(ctx context.Context) error {
	return cmd.ExecuteContext(ctx)
}

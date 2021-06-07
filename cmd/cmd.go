package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/osixia/container-baseimage/cmd/entrypoint"
	"github.com/osixia/container-baseimage/cmd/logger"
	"github.com/osixia/container-baseimage/cmd/services"
)

const (
	installGroupID      = "install"
	entrypointGroupID   = "entrypoint"
	servicesGroupID     = "services"
	distributionGroupID = "distribution"
	envsubstGroupID     = "envsubst"
	loggerGroupID       = "logger"
)

var cmd = &cobra.Command{
	Use: "container-baseimage",

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return logger.HandleFlags(cmd)
	},
}

func init() {
	cobra.EnableCommandSorting = false

	// subcommands groups
	cmd.AddGroup(&cobra.Group{
		ID:    installGroupID,
		Title: "Install Command:",
	})

	cmd.AddGroup(&cobra.Group{
		ID:    entrypointGroupID,
		Title: "Entrypoint Command:",
	})

	cmd.AddGroup(&cobra.Group{
		ID:    servicesGroupID,
		Title: "Services Commands:",
	})

	cmd.AddGroup(&cobra.Group{
		ID:    distributionGroupID,
		Title: "Distribution Commands:",
	})

	cmd.AddGroup(&cobra.Group{
		ID:    envsubstGroupID,
		Title: "Envsubst Commands:",
	})

	cmd.AddGroup(&cobra.Group{
		ID:    loggerGroupID,
		Title: "Logger Command:",
	})

	// subcommands
	entrypoint.EntrypointCmd.GroupID = entrypointGroupID
	cmd.AddCommand(entrypoint.EntrypointCmd)

	services.ServicesCmd.GroupID = servicesGroupID
	cmd.AddCommand(services.ServicesCmd)

	cmd.AddCommand(installCmd)
	cmd.AddCommand(addMultiprocessStackCmd)
	cmd.AddCommand(installDebugPackagesCmd)

	cmd.AddCommand(envsubstCmd)
	cmd.AddCommand(envsubstTemplatesCmd)

	logger.LoggerCmd.GroupID = loggerGroupID
	cmd.AddCommand(logger.LoggerCmd)
}

func Run(ctx context.Context) error {
	return cmd.ExecuteContext(ctx)
}

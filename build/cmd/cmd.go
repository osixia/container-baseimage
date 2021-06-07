package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/osixia/container-baseimage/build/cmd/job"
	"github.com/osixia/container-baseimage/build/cmd/provider"
)

const (
	cicdGroupID = "cicd"
)

var cmd = &cobra.Command{
	Use: "build",
}

func init() {
	cobra.EnableCommandSorting = false

	// subcommands groups
	cmd.AddGroup(&cobra.Group{
		ID:    cicdGroupID,
		Title: "CI/CD Commands:",
	})

	// subcommands
	provider.ProviderCmd.GroupID = cicdGroupID
	job.JobCmd.GroupID = cicdGroupID
	cmd.AddCommand(provider.ProviderCmd)
	cmd.AddCommand(job.JobCmd)
}

func Run(ctx context.Context) error {
	return cmd.ExecuteContext(ctx)
}

package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/osixia/container-baseimage/ci/config"
	"github.com/osixia/container-baseimage/ci/job"
)

type jobFlags interface {
	SetDockerfile(d string)
	SetVersion(v string)

	toJobOptions() interface{}
}

const (
	pipelineGroupID = "pipeline"
	jobsGroupID     = "jobs"
)

var cmd = &cobra.Command{
	Use: "ci",
}

func init() {
	cobra.EnableCommandSorting = false

	// subcommands groups
	cmd.AddGroup(&cobra.Group{
		ID:    pipelineGroupID,
		Title: "Pipeline:",
	})

	cmd.AddGroup(&cobra.Group{
		ID:    jobsGroupID,
		Title: "Jobs:",
	})

	// subcommands
	cmd.AddCommand(githubCmd)

	cmd.AddCommand(buildCmd)
	cmd.AddCommand(testCmd)
	cmd.AddCommand(deployCmd)
}

func Run(ctx context.Context) error {
	return cmd.ExecuteContext(ctx)
}

func newStepCmd(use string, short string, alias string, j job.Job, cf jobFlags) *cobra.Command {

	return &cobra.Command{

		Use:   fmt.Sprintf("%v dockerfile [version]", use),
		Short: short,

		GroupID: jobsGroupID,

		Aliases: []string{
			alias,
		},

		Args: cobra.RangeArgs(1, 2),

		Run: func(cmd *cobra.Command, args []string) {

			cf.SetDockerfile(args[0])

			cf.SetVersion(config.DefaultVersion)

			if len(args) > 1 {
				cf.SetVersion(args[1])
			}

			if err := job.Run(j, cmd.Context(), cf.toJobOptions()); err != nil {
				fatal(err)
			}
		},
	}
}

func fatal(err error) {
	fmt.Printf("error: %v\n", err.Error())
	os.Exit(1)
}

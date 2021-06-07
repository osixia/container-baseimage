package helpers

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/osixia/container-baseimage/build/job"
)

type JobFlags interface {
	SetDockerfile(df string)

	ToJobOptions() any
}

func NewJobCmd(use string, short string, alias string, j job.JobFunc, jf JobFlags) *cobra.Command {

	return &cobra.Command{

		Use:   fmt.Sprintf("%v dockerfile", use),
		Short: short,

		Aliases: []string{
			alias,
		},

		Args: cobra.ExactArgs(1),

		Run: func(cmd *cobra.Command, args []string) {

			jf.SetDockerfile(args[0])

			if _, err := job.Run(cmd.Context(), j, jf.ToJobOptions()); err != nil {
				log.Fatal(err)
			}
		},
	}
}

package helpers

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/osixia/container-baseimage/log"
)

func NewPrintCmd(use string, short string, alias string, f func() string) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,

		Aliases: []string{
			alias,
		},

		Args: cobra.NoArgs,

		Run: func(cmd *cobra.Command, args []string) {
			log.Tracef("Run: %v called with args: %v", cmd.Use, args)

			fmt.Println(f())
		},
	}
}

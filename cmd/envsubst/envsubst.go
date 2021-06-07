package envsubst

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	cmdlog "github.com/osixia/container-baseimage/cmd/log"
	"github.com/osixia/container-baseimage/helpers"
	"github.com/osixia/container-baseimage/log"
)

var EnvsubstCmd = &cobra.Command{
	Use:   "envsubst string [string]...",
	Short: "Template rendering with environment variables",

	Args: cobra.MinimumNArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		s := strings.Join(args, " ")

		r, err := helpers.Envsubst(s)
		if err != nil {
			log.Fatalf("%v: %v", cmd.Use, err.Error())
		}

		fmt.Println(r)
	},
}

func init() {
	// subcommands
	EnvsubstCmd.AddCommand(fileCmd)
	EnvsubstCmd.AddCommand(templatesCmd)

	// flags
	EnvsubstCmd.Flags().SortFlags = false
	cmdlog.AddFlags(EnvsubstCmd.Flags())
}

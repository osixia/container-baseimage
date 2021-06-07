package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/osixia/container-baseimage/cmd/logger"
	"github.com/osixia/container-baseimage/helpers"
	"github.com/osixia/container-baseimage/log"
)

var envsubstCmd = &cobra.Command{
	Use:   "envsubst string [string]...",
	Short: "Envsubst",

	GroupID: envGroupID,

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
	// flags
	envsubstCmd.Flags().SortFlags = false
	logger.AddFlags(envsubstCmd.Flags())
}

package cmd

import (
	"github.com/spf13/cobra"

	"github.com/osixia/container-baseimage/cmd/logger"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/log"
)

var addMultiprocessStackCmd = &cobra.Command{
	Use:   "add-multiprocess-stack",
	Short: "Add multiprocess stack to container service(s)",

	GroupID: distributionGroupID,

	Aliases: []string{
		"ams",
	},

	Args: cobra.NoArgs,

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		if err := core.Instance().Distribution().AddMultiprocessStack(cmd.Context()); err != nil {
			log.Fatalf("%v: %v", cmd.Use, err.Error())
		}
	},
}

func init() {
	// flags
	addMultiprocessStackCmd.Flags().SortFlags = false
	logger.AddFlags(addMultiprocessStackCmd.Flags())
}

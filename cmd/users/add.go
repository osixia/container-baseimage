package users

import (
	"github.com/spf13/cobra"

	"github.com/osixia/container-baseimage/cmd/logger"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/log"
)

type addFlags struct {
	groupID   string
	groupName string
}

var addCmdFlags = &addFlags{}

var addCmd = &cobra.Command{
	Use:   "add id name --group-id group_id --group-name group_name",
	Short: "Add user",

	Args: cobra.ExactArgs(2),

	Aliases: []string{
		"a",
	},

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		if err := core.Instance().Distribution().AddUser(cmd.Context(), args[0], args[1], addCmdFlags.groupID, addCmdFlags.groupName); err != nil {
			log.Fatalf("%v: %v", cmd.Use, err.Error())
		}
	},
}

func init() {
	// flags
	addCmd.Flags().SortFlags = false

	addCmd.Flags().StringVarP(&addCmdFlags.groupID, "group-id", "g", "", "group id")
	addCmd.Flags().StringVarP(&addCmdFlags.groupName, "group-name", "i", "", "group name\n")
	if err := addCmd.MarkFlagRequired("group-id"); err != nil {
		log.Fatal(err.Error())
	}
	if err := addCmd.MarkFlagRequired("group-name"); err != nil {
		log.Fatal(err.Error())
	}

	logger.AddFlags(addCmd.Flags())
}

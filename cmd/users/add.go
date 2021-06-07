package users

import (
	"github.com/spf13/cobra"

	cmdlog "github.com/osixia/container-baseimage/cmd/log"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/helpers"
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

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		helpers.Must(core.Instance().Distribution().AddUser(cmd.Context(), args[0], args[1], addCmdFlags.groupID, addCmdFlags.groupName))
	},
}

func init() {
	// flags
	addCmd.Flags().SortFlags = false

	addCmd.Flags().StringVarP(&addCmdFlags.groupID, "group-id", "i", "", "group id")
	addCmd.Flags().StringVarP(&addCmdFlags.groupName, "group-name", "n", "", "group name\n")
	helpers.Must(addCmd.MarkFlagRequired("group-id"))
	helpers.Must(addCmd.MarkFlagRequired("group-name"))

	cmdlog.AddFlags(addCmd.Flags())
}

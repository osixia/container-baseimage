package users

import (
	"github.com/spf13/cobra"
)

var UsersCmd = &cobra.Command{
	Use:   "users",
	Short: "Manage users",
}

func init() {
	//subcommands
	UsersCmd.AddCommand(addCmd)
}

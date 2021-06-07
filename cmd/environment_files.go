package cmd

import (
	"strings"

	"github.com/osixia/container-baseimage/cmd/logger"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/helpers"
	"github.com/osixia/container-baseimage/log"
)

var environmentFilesFunc = func() string {
	efs, err := core.Instance().Filesystem().ListDotEnv()
	if err != nil {
		log.Fatal(err.Error())
	}
	return strings.Join(efs, "\n")
}

var environmentFilesCmd = helpers.NewPrintCmd("environment-files", "List environment files", "env", environmentFilesFunc)

func init() {

	environmentFilesCmd.GroupID = envGroupID

	// flags
	environmentFilesCmd.Flags().SortFlags = false
	logger.AddFlags(environmentFilesCmd.Flags())
}

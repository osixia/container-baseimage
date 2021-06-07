package environment

import (
	"strings"

	"github.com/osixia/container-baseimage/cmd/logger"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/helpers"
	"github.com/osixia/container-baseimage/log"
)

var filesFunc = func() string {
	efs, err := core.Instance().Filesystem().ListDotEnv()
	if err != nil {
		log.Fatal(err.Error())
	}
	return strings.Join(efs, "\n")
}

var filesCmd = helpers.NewPrintCmd("files", "List environment files", []string{"f"}, filesFunc)

func init() {
	// flags
	filesCmd.Flags().SortFlags = false
	logger.AddFlags(filesCmd.Flags())
}

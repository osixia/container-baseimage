package environment

import (
	"strings"

	cmdlog "github.com/osixia/container-baseimage/cmd/log"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/helpers"
)

var filesFunc = func() string {
	efs := helpers.MustVal(core.Instance().Filesystem().ListDotEnv())
	return strings.Join(efs, "\n")
}

var filesCmd = helpers.NewPrintCmd("files", "List environment files", []string{"f"}, filesFunc)

func init() {
	// flags
	filesCmd.Flags().SortFlags = false
	cmdlog.AddFlags(filesCmd.Flags())
}

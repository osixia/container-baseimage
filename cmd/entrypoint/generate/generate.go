package generate

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/log"
)

type generateFlags struct {
	print bool
}

const (
	Separator = "--------------------------------------------------------------------------------"
)

var GenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate sample templates",

	Aliases: []string{
		"gen",
	},
}

func init() {
	// subcommands
	GenerateCmd.AddCommand(bootstrapCmd)
	GenerateCmd.AddCommand(dockerfileCmd)
	GenerateCmd.AddCommand(environmentCmd)
	GenerateCmd.AddCommand(servicesCmd)
}

func addGenerateFlags(fs *pflag.FlagSet, gopt *generateFlags) {
	fs.BoolVar(&gopt.print, "print", false, "print generated files content\n")
}

func print(files []string) {

	dirPrefix := core.Instance().Filesystem().Paths().RunGenerator

	for _, f := range files {
		c, err := os.ReadFile(f)
		if err != nil {
			log.Fatal(err.Error())
		}

		fmt.Printf("\n%v\n%v\n%v%v\n", strings.TrimPrefix(f, dirPrefix), Separator, string(c), Separator)
	}

}

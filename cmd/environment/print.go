package environment

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/osixia/container-baseimage/cmd/logger"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/log"
)

var printCleanVars = []string{"HOSTNAME", "PWD", "HOME", "LS_COLORS", "TERM", "SHLVL", "PATH", "_"}

type printFlags struct {
	shell bool

	exclude []string
	clean   bool
}

var printCmdFlags = &printFlags{}

var printCmd = &cobra.Command{
	Use:   "print [variable]...",
	Short: "Print environment",

	Aliases: []string{
		"p",
	},

	Run: func(cmd *cobra.Command, args []string) {
		log.Tracef("Run: %v called with args: %v", cmd.Use, args)

		if printCmdFlags.clean {
			printCmdFlags.exclude = append(printCmdFlags.exclude, printCleanVars...)
		}

		fmt.Print(core.Instance().Environment().Print(printCmdFlags.shell, args, printCmdFlags.exclude))
	},
}

func init() {
	// flags
	printCmd.Flags().SortFlags = false

	printCmd.Flags().BoolVarP(&printCmdFlags.shell, "shell", "s", false, "print environment as shell variables\n")
	printCmd.Flags().StringArrayVarP(&printCmdFlags.exclude, "exclude", "e", nil, "exclude environment variable")
	printCmd.Flags().BoolVarP(&printCmdFlags.clean, "clean", "c", false, fmt.Sprintf("exclude standard environment variables: %v\n", strings.Join(printCleanVars, ", ")))

	logger.AddFlags(printCmd.Flags())
}

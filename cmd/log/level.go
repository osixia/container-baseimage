package log

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/osixia/container-baseimage/helpers"
	"github.com/osixia/container-baseimage/log"
)

const comparisonGroupID = "comparison"

var levelFunc = func() string {
	return log.Levels[log.Level()]
}

var levelCmd = helpers.NewPrintCmd("level", "", []string{"lvl"}, levelFunc)

func init() {
	// subcommands groups
	levelCmd.AddGroup(&cobra.Group{
		ID:    comparisonGroupID,
		Title: "Comparison Commands:",
	})

	// subcommands
	levelCmd.AddCommand(newLevelCompareCmd("eq", "Equals", log.Equals))
	levelCmd.AddCommand(newLevelCompareCmd("ne", "Not equals", log.NotEquals))
	levelCmd.AddCommand(newLevelCompareCmd("gt", "Greater than", log.GreaterThan))
	levelCmd.AddCommand(newLevelCompareCmd("ge", "Greater or equals", log.GreaterOrEquals))
	levelCmd.AddCommand(newLevelCompareCmd("lt", "Less than", log.LessThan))
	levelCmd.AddCommand(newLevelCompareCmd("le", "Less or equals", log.LessOrEquals))
}

func newLevelCompareCmd(use string, short string, f log.CompareFunc) *cobra.Command {

	return &cobra.Command{
		Use:   fmt.Sprintf("%v [%v]", use, strings.Join(log.LevelsList(), ",")),
		Short: short,

		GroupID: comparisonGroupID,

		Args: cobra.ExactArgs(1),

		RunE: func(cmd *cobra.Command, args []string) error {

			level, err := log.ParseLevel(args[0])
			if err != nil {
				return err
			}

			if f(log.Level(), level) {
				os.Exit(0)
			}

			os.Exit(1)
			return nil // unreachable, satisfies compiler
		},
	}
}

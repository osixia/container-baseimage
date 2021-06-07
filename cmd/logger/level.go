package logger

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/osixia/container-baseimage/log"
)

var levelCmd = &cobra.Command{
	Use: "level",
}

func init() {
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

		Args: cobra.ExactArgs(1),

		RunE: func(cmd *cobra.Command, args []string) error {

			level, err := log.ParseLevel(args[0])
			if err != nil {
				return err
			}

			if f(log.Level(), level) {
				os.Exit(0)
				return nil
			}

			os.Exit(1)
			return nil
		},
	}
}

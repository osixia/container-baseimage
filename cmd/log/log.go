package log

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/osixia/container-baseimage/log"
)

const (
	LevelFlag  = "log-level"
	QuietFlag  = "quiet"
	FormatFlag = "log-format"

	PrintGroupID = "print"
)

var caser = cases.Title(language.English)

var LogCmd = &cobra.Command{
	Use:   "log",
	Short: "Logging utilities",

	Aliases: []string{
		"l",
	},
}

func init() {
	// subcommands groups
	LogCmd.AddGroup(&cobra.Group{
		ID:    PrintGroupID,
		Title: "Print Commands:",
	})

	// subcommands
	LogCmd.AddCommand(newPrintLogCmd("fatal", nil, log.Fatal))
	LogCmd.AddCommand(newPrintLogCmd(log.Levels[log.LevelError], []string{"err"}, log.Error))
	LogCmd.AddCommand(newPrintLogCmd(log.Levels[log.LevelWarning], []string{"warn"}, log.Warning))
	LogCmd.AddCommand(newPrintLogCmd(log.Levels[log.LevelInfo], nil, log.Info))
	LogCmd.AddCommand(newPrintLogCmd(log.Levels[log.LevelDebug], []string{"dbg"}, log.Debug))
	LogCmd.AddCommand(newPrintLogCmd(log.Levels[log.LevelTrace], []string{"trc"}, log.Trace))

	LogCmd.AddCommand(levelCmd)
}

func AddFlags(fs *pflag.FlagSet) {
	fs.StringP(LevelFlag, "l", log.Levels[log.Level()], fmt.Sprintf("set log level, choices: %v", strings.Join(log.LevelsList(), ", ")))
	fs.Bool(QuietFlag, false, "show errors only")
	fs.StringP(FormatFlag, "o", string(log.Format()), fmt.Sprintf("set log format, choices: %v", strings.Join(log.FormatsList(), ", ")))
}

func HandleFlags(cmd *cobra.Command) error {

	level, err := cmd.Flags().GetString(LevelFlag)
	if err == nil {
		if err := log.SetLevel(level); err != nil {
			return err
		}
	}

	quiet, err := cmd.Flags().GetBool(QuietFlag)
	if err == nil && quiet {
		if err := log.SetLevel(log.Levels[log.LevelError]); err != nil {
			return err
		}
	}

	format, err := cmd.Flags().GetString(FormatFlag)
	if err == nil {
		if err := log.SetFormat(format); err != nil {
			return err
		}
	}

	return nil
}

func newPrintLogCmd(level string, alias []string, f log.PrintFunc) *cobra.Command {

	return &cobra.Command{
		Use:   fmt.Sprintf("%v message", level),
		Short: caser.String(level),

		GroupID: PrintGroupID,

		Aliases: alias,

		Run: func(cmd *cobra.Command, args []string) {
			log.FromCmd(f, args)
		},
	}
}

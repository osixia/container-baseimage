package logger

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
	FormatFlag = "log-format"

	PrintGroupID = "print"
)

var caser = cases.Title(language.English)

var LoggerCmd = &cobra.Command{
	Use:   "logger",
	Short: "Logger subcommands",

	Aliases: []string{
		"l",
	},
}

func init() {
	// subcommands groups
	LoggerCmd.AddGroup(&cobra.Group{
		ID:    PrintGroupID,
		Title: "Print Commands:",
	})

	// subcommands
	LoggerCmd.AddCommand(newPrintCmd(log.Levels[log.LevelError], "e", log.Error))
	LoggerCmd.AddCommand(newPrintCmd(log.Levels[log.LevelWarning], "w", log.Warning))
	LoggerCmd.AddCommand(newPrintCmd(log.Levels[log.LevelInfo], "i", log.Info))
	LoggerCmd.AddCommand(newPrintCmd(log.Levels[log.LevelDebug], "d", log.Debug))
	LoggerCmd.AddCommand(newPrintCmd(log.Levels[log.LevelTrace], "t", log.Trace))

	LoggerCmd.AddCommand(levelCmd)
}

func AddFlags(fs *pflag.FlagSet) {
	fs.StringP(LevelFlag, "l", log.Levels[log.Level()], fmt.Sprintf("set log level, choices: %v", strings.Join(log.LevelsList(), ", ")))
	fs.StringP(FormatFlag, "o", string(log.Format()), fmt.Sprintf("set log format, choices: %v", strings.Join(log.FormatsList(), ", ")))
}

func HandleFlags(cmd *cobra.Command) error {

	level, err := cmd.Flags().GetString(LevelFlag)
	if err == nil {
		if err := log.SetLevel(level); err != nil {
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

func newPrintCmd(level string, alias string, f log.PrintFunc) *cobra.Command {

	return &cobra.Command{
		Use:   fmt.Sprintf("%v message", level),
		Short: caser.String(level),

		GroupID: PrintGroupID,

		Aliases: []string{
			alias,
		},

		Run: func(cmd *cobra.Command, args []string) {
			log.FromCmd(f, args)
		},
	}
}

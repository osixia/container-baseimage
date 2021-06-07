package log

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"
)

type level uint32
type format string

type PrintFunc func(message string)

type CompareFunc func(level, level) bool

const (
	LevelNone    level = 0
	LevelError   level = 1
	LevelWarning level = 2
	LevelInfo    level = 3
	LevelDebug   level = 4
	LevelTrace   level = 5

	FormatConsole format = "console"
	FormatJson    format = "json"
)

var ErrBadLevel = errors.New("invalid log level")
var ErrBadFormat = errors.New("invalid log format")
var ErrRequired = errors.New("required")

var Levels = map[level]string{
	LevelNone:    "none",
	LevelError:   "error",
	LevelWarning: "warning",
	LevelInfo:    "info",
	LevelDebug:   "debug",
	LevelTrace:   "trace",
}

var Formats = map[format]string{
	FormatConsole: "console",
	FormatJson:    "json",
}

var DefaultLevel = LevelInfo
var DefaultFormat = FormatConsole

type Config struct {
	Level  level
	Format format
}

var config = &Config{
	Level:  DefaultLevel,
	Format: DefaultFormat,
}

type EnvironmentConfig struct {
	LevelKey  string
	FormatKey string
}

var environmentConfig *EnvironmentConfig

func (ec *EnvironmentConfig) Validate() error {
	if ec.LevelKey == "" {
		return fmt.Errorf("%v: %w", "LevelKey", ErrRequired)
	}

	if ec.FormatKey == "" {
		return fmt.Errorf("%v: %w", "FormatKey", ErrRequired)
	}

	return nil
}

func SetEnvironmentConfig(ec *EnvironmentConfig) error {

	if err := ec.Validate(); err != nil {
		return err
	}

	environmentConfig = ec

	if os.Getenv(ec.LevelKey) != "" {
		if err := SetLevel(os.Getenv(ec.LevelKey)); err != nil {
			return err
		}
	}

	if os.Getenv(ec.FormatKey) != "" {
		if err := SetFormat(os.Getenv(ec.FormatKey)); err != nil {
			return err
		}
	}

	return nil
}

func SetConfig(c *Config) {
	config = c
}

func FromCmd(f PrintFunc, args []string) {
	fi, _ := os.Stdin.Stat()
	if (fi.Mode() & os.ModeCharDevice) == 0 {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			f(scanner.Text())
		}
	} else {
		f(strings.Join(args, " "))
	}
}

func print(output io.Writer, level string, message string) {
	now := time.Now().Format(time.RFC3339)

	if config.Format == FormatJson {
		dictionary := map[string]string{
			"datetime": now,
			"level":    strings.TrimSpace(level),
			"message":  message,
		}

		json, err := json.Marshal(dictionary)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}

		fmt.Fprintln(output, string(json))
		return
	}

	for _, line := range strings.Split(message, "\n") {
		fmt.Fprintf(output, "%v %v %v\n", now, level, line)
	}
}

func Fatal(message string) {
	Error(message)
	os.Exit(1)
}

func Fatalf(format string, a ...any) {
	Fatal(fmt.Sprintf(format, a...))
}

func Error(message string) {
	if config.Level >= LevelError {
		print(os.Stderr, "ERROR  ", message)
	}
}

func Errorf(format string, a ...any) {
	Error(fmt.Sprintf(format, a...))
}

func Warning(message string) {
	if config.Level >= LevelWarning {
		print(os.Stdout, "WARNING", message)
	}
}

func Warningf(format string, a ...any) {
	Warning(fmt.Sprintf(format, a...))
}

func Info(message string) {
	if config.Level >= LevelInfo {
		print(os.Stdout, "INFO   ", message)
	}
}

func Infof(format string, a ...any) {
	Info(fmt.Sprintf(format, a...))
}

func Debug(message string) {
	if config.Level >= LevelDebug {
		print(os.Stdout, "DEBUG  ", message)
	}
}

func Debugf(format string, a ...any) {
	Debug(fmt.Sprintf(format, a...))
}

func Trace(message string) {
	if config.Level >= LevelTrace {
		print(os.Stdout, "TRACE  ", message)
	}
}

func Tracef(format string, a ...any) {
	Trace(fmt.Sprintf(format, a...))
}

func LevelsList() []string {
	keys := make([]int, 0, len(Levels))
	for k := range Levels {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)

	values := make([]string, 0, len(Levels))
	for k := range keys {
		values = append(values, Levels[level(k)])
	}

	return values
}

func FormatsList() []string {
	values := make([]string, 0, len(Formats))
	for _, f := range Formats {
		values = append(values, f)
	}
	sort.Strings(values)

	return values
}

func ParseLevel(level string) (level, error) {
	for k, l := range Levels {
		if l == level {
			return k, nil
		}
	}

	return 0, fmt.Errorf("%v: %w (choices: %v)", level, ErrBadLevel, strings.Join(LevelsList(), ", "))
}

func Level() level {
	return config.Level
}

func SetLevel(level string) error {
	lvl, err := ParseLevel(level)
	if err != nil {
		return err
	}

	config.Level = lvl

	if environmentConfig != nil {
		os.Setenv(environmentConfig.LevelKey, level)
	}

	return nil
}

func Format() format {
	return config.Format
}

func ParseFormat(format string) (format, error) {
	for k, f := range Formats {
		if f == format {
			return k, nil
		}
	}

	return "", fmt.Errorf("%v: %w", format, ErrBadFormat)
}

func SetFormat(format string) error {
	f, err := ParseFormat(format)
	if err != nil {
		return err
	}

	config.Format = f

	if environmentConfig != nil {
		os.Setenv(environmentConfig.FormatKey, Formats[config.Format])
	}

	return nil
}

func Equals(a level, b level) bool          { return a == b }
func NotEquals(a level, b level) bool       { return a != b }
func GreaterThan(a level, b level) bool     { return a > b }
func GreaterOrEquals(a level, b level) bool { return a >= b }
func LessThan(a level, b level) bool        { return a < b }
func LessOrEquals(a level, b level) bool    { return a <= b }

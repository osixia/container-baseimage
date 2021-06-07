package log

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/osixia/container-baseimage/errors"
)

type level uint32
type format string

type PrintFunc func(message string)
type CompareFunc func(level, level) bool

const (
	LevelError   level = 0
	LevelWarning level = 1
	LevelInfo    level = 2
	LevelDebug   level = 3
	LevelTrace   level = 4

	FormatConsole format = "console"
	FormatJson    format = "json"
)

var Levels = map[level]string{
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

var stdout io.Writer = os.Stdout
var stderr io.Writer = os.Stderr

// Logger uses the global level and format with an optional output override.
// Its zero value preserves the default stdout/stderr routing.
type Logger struct {
	output io.Writer
}

var defaultLogger = &Logger{}

// WithOutput selects an output without changing the default logger.
// A nil output preserves the default routing.
func WithOutput(output io.Writer) *Logger {
	return &Logger{output: output}
}

var (
	Fatal    = defaultLogger.Fatal
	Fatalf   = defaultLogger.Fatalf
	Error    = defaultLogger.Error
	Errorf   = defaultLogger.Errorf
	Warning  = defaultLogger.Warning
	Warningf = defaultLogger.Warningf
	Info     = defaultLogger.Info
	Infof    = defaultLogger.Infof
	Debug    = defaultLogger.Debug
	Debugf   = defaultLogger.Debugf
	Trace    = defaultLogger.Trace
	Tracef   = defaultLogger.Tracef
)

var environmentConfig *EnvironmentConfig

var consoleLabels = map[level]string{
	LevelError:   "ERROR  ",
	LevelWarning: "WARNING",
	LevelInfo:    "INFO   ",
	LevelDebug:   "DEBUG  ",
	LevelTrace:   "TRACE  ",
}

func (ec *EnvironmentConfig) Validate() (bool, error) {

	if ec.LevelKey == "" {
		return false, fmt.Errorf("LevelKey: %w", errors.ErrRequired)
	}

	if ec.FormatKey == "" {
		return false, fmt.Errorf("FormatKey: %w", errors.ErrRequired)
	}

	return true, nil
}

func SetEnvironmentConfig(ec *EnvironmentConfig) error {

	if _, err := ec.Validate(); err != nil {
		return err
	}

	environmentConfig = ec

	if value := os.Getenv(ec.LevelKey); value != "" {
		if err := SetLevel(value); err != nil {
			return err
		}
	}

	if value := os.Getenv(ec.FormatKey); value != "" {
		if err := SetFormat(value); err != nil {
			return err
		}
	}

	return nil
}

func SetConfig(c *Config) {
	config = c
}

func FromCmd(f PrintFunc, args []string) {

	if len(args) > 0 {
		f(strings.Join(args, " "))
		return
	}

	fi, err := os.Stdin.Stat()
	if err != nil {
		Fatalf("Error reading stdin: %v", err)
		return
	}

	if (fi.Mode() & os.ModeCharDevice) == 0 {
		writer := NewWriter(f)
		_, err := io.Copy(writer, os.Stdin)
		writer.Flush()
		if err != nil {
			Fatalf("Error reading stdin: %v", err)
		}
	}
}

func (l *Logger) Fatal(message string) {

	l.Error(message)
	os.Exit(1)
}

func (l *Logger) Fatalf(format string, a ...any) {

	l.Errorf(format, a...)
	os.Exit(1)
}

func (l *Logger) Error(message string) {
	l.log(LevelError, message)
}

func (l *Logger) Errorf(format string, a ...any) {
	l.log(LevelError, format, a...)
}

func (l *Logger) Warning(message string) {
	l.log(LevelWarning, message)
}

func (l *Logger) Warningf(format string, a ...any) {
	l.log(LevelWarning, format, a...)
}

func (l *Logger) Info(message string) {
	l.log(LevelInfo, message)
}

func (l *Logger) Infof(format string, a ...any) {
	l.log(LevelInfo, format, a...)
}

func (l *Logger) Debug(message string) {
	l.log(LevelDebug, message)
}

func (l *Logger) Debugf(format string, a ...any) {
	l.log(LevelDebug, format, a...)
}

func (l *Logger) Trace(message string) {
	l.log(LevelTrace, message)
}

func (l *Logger) Tracef(format string, a ...any) {
	l.log(LevelTrace, format, a...)
}

func LevelsList() []string {

	keys := make([]int, 0, len(Levels))
	for k := range Levels {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)

	values := make([]string, 0, len(Levels))
	for _, k := range keys {
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

	return 0, fmt.Errorf("%v: log level %w (choices: %v)", level, errors.ErrUnknown, strings.Join(LevelsList(), ", "))
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
		if err := os.Setenv(environmentConfig.LevelKey, level); err != nil {
			return fmt.Errorf("failed to set environment variable %v=%v", environmentConfig.LevelKey, level)
		}
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

	return "", fmt.Errorf("%v: log format %w (choices: %v)", format, errors.ErrUnknown, strings.Join(FormatsList(), ", "))
}

func SetFormat(format string) error {

	f, err := ParseFormat(format)
	if err != nil {
		return err
	}

	config.Format = f

	if environmentConfig != nil {
		if err := os.Setenv(environmentConfig.FormatKey, Formats[config.Format]); err != nil {
			return fmt.Errorf("failed to set environment variable %v=%v", environmentConfig.FormatKey, Formats[config.Format])
		}
	}

	return nil
}

func Equals(a level, b level) bool          { return a == b }
func NotEquals(a level, b level) bool       { return a != b }
func GreaterThan(a level, b level) bool     { return a > b }
func GreaterOrEquals(a level, b level) bool { return a >= b }
func LessThan(a level, b level) bool        { return a < b }
func LessOrEquals(a level, b level) bool    { return a <= b }

func (logger *Logger) log(l level, message string, a ...any) {

	if config.Level < l {
		return
	}

	output := logger.output
	if output == nil {
		output = stdout
		if l == LevelError {
			output = stderr
		}
	}
	print(output, l, getMessage(message, a...))
}

func print(output io.Writer, l level, message string) {

	now := time.Now().Format(time.RFC3339)

	if config.Format == FormatJson {
		dictionary := map[string]string{
			"datetime": now,
			"level":    Levels[l],
			"message":  message,
		}

		data, err := json.Marshal(dictionary)
		if err != nil {
			_, _ = fmt.Fprintln(stderr, err)
			return
		}

		_, _ = fmt.Fprintln(output, string(data))
		return
	}

	label := consoleLabels[l]
	for _, line := range strings.Split(message, "\n") {
		_, _ = fmt.Fprintf(output, "%v %v %v\n", now, label, line)
	}
}

// getMessage formats with Sprint, Sprintf, or neither.
func getMessage(format string, a ...any) string {

	if len(a) == 0 {
		return format
	}

	if format != "" {
		return fmt.Sprintf(format, a...)
	}

	if len(a) == 1 {
		if str, ok := a[0].(string); ok {
			return str
		}
	}

	return fmt.Sprint(a...)
}

package errors

import (
	"errors"
	"os/exec"
)

// Errors global variables
// =============================

var ErrRequired = errors.New("required")
var ErrUnknown = errors.New("unknown")
var ErrUnavailable = errors.New("unavailable")

func ExitCode(err error) int {
	if err == nil {
		return 0
	}

	exitCode := 1

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		exitCode = exitErr.ExitCode()
	}

	return exitCode
}

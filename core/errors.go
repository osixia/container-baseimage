package core

import (
	"errors"
	"os/exec"
)

// generic errors
var ErrRequired = errors.New("required data missing")
var ErrEnvironmentVariableRequired = errors.New("environment variable required")

// service errors
var ErrServiceNotFound = errors.New("service not found")
var ErrLinkedServiceNotFound = errors.New("linked service not found")
var ErrServiceExists = errors.New("service already exists")
var ErrMultiprocessStackAlreadyAdded = errors.New("multiprocess stack already added")
var ErrMultiprocessStackPartiallyAdded = errors.New("missing multiprocess stack service(s) (not installed or not linked to entrypoint)")

// entrypoint error
var ErrNotSingleProcessContainer = errors.New("not a single process container")

// distribution errors
var ErrDistributionNotFound = errors.New("linux distribution not found")
var ErrDistributionNotSupported = errors.New("linux distribution not supported")

// linked service error
var ErrLinkedScriptNotFound = errors.New("script not found")

func IsErrServiceNotFound(err error) bool {
	return err == ErrServiceNotFound || errors.Unwrap(err) == ErrServiceNotFound
}

func IsErrLinkedServiceNotFound(err error) bool {
	return err == ErrLinkedServiceNotFound || errors.Unwrap(err) == ErrLinkedServiceNotFound
}

func IsExitError(err error) bool {
	var eerr *exec.ExitError
	return errors.As(err, &eerr)
}

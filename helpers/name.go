package helpers

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/osixia/container-baseimage/log"
)

// ValidateName checks a service, process or tag is a single path component.
func ValidateName(name string) error {

	log.Tracef("ValidateName called with name: %v", name)

	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, "/\\\x00") {
		return fmt.Errorf("invalid name %q: expected a single path component", name)
	}

	return nil
}

// ValidateNames checks each name and returns the first validation error.
func ValidateNames(names ...string) error {

	log.Tracef("ValidateNames called with names: %v", names)

	for _, name := range names {
		if err := ValidateName(name); err != nil {
			return err
		}
	}

	return nil
}

var unixIDRegexp = regexp.MustCompile(`^[0-9]+$`)

// ValidateUnixID checks a string is a plain numeric user/group id, safe to
// interpolate into a system-command template without argument injection.
func ValidateUnixID(id string) error {

	log.Tracef("ValidateUnixID called with id: %v", id)

	if !unixIDRegexp.MatchString(id) {
		return fmt.Errorf("invalid id %q: expected digits only", id)
	}

	return nil
}

var unixNameRegexp = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_-]*$`)

// ValidateUnixName checks a string is a safe POSIX user/group name, rejecting
// whitespace, leading dashes and anything else that could be parsed as a
// separate shell word or flag once interpolated into a system-command template.
func ValidateUnixName(name string) error {

	log.Tracef("ValidateUnixName called with name: %v", name)

	if !unixNameRegexp.MatchString(name) {
		return fmt.Errorf("invalid name %q: expected a POSIX user/group name", name)
	}

	return nil
}

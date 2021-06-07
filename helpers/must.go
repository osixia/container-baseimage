package helpers

import (
	"fmt"

	"github.com/osixia/container-baseimage/log"
)

func mustFatal(err error, format string, args ...any) {
	if err == nil {
		return
	}

	if format != "" {
		log.Fatalf("%s: %v", fmt.Sprintf(format, args...), err)
		return
	}

	log.Fatal(err.Error())
}

func Must(err error) {
	mustFatal(err, "")
}

func Mustf(err error, format string, args ...any) {
	mustFatal(err, format, args...)
}

func MustVal[T any](v T, err error) T {
	mustFatal(err, "")
	return v
}

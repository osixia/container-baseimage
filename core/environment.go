package core

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/osixia/container-baseimage/errors"
	"github.com/osixia/container-baseimage/helpers"
)

// Environment config
// =============================

type EnvironmentConfig struct {
	ImageKey         string
	DebugPackagesKey string
}

func (envc *EnvironmentConfig) Validate() (bool, error) {

	if envc.ImageKey == "" {
		return false, fmt.Errorf("ImageKey: %w", errors.ErrRequired)
	}

	if envc.DebugPackagesKey == "" {
		return false, fmt.Errorf("DebugPackagesKey: %w", errors.ErrRequired)
	}

	return true, nil
}

// Environment
// =============================

type Environment interface {
	Print(shell bool, include, exclude []string) string

	Image() string
	DebugPackages() string

	Config() *EnvironmentConfig
}

type environment struct {
	config *EnvironmentConfig
}

func newEnvironment(envc *EnvironmentConfig) (Environment, error) {

	if _, err := envc.Validate(); err != nil {
		return nil, err
	}

	return &environment{
		config: envc,
	}, nil
}

func (env *environment) Print(shell bool, include, exclude []string) string {

	var s strings.Builder

	for _, e := range os.Environ() {

		kv := strings.SplitN(e, "=", 2)
		k := kv[0]
		v := ""

		if len(kv) > 1 {
			v = kv[1]
		}

		if len(include) > 0 && !slices.Contains(include, k) {
			continue
		}

		if len(exclude) > 0 && slices.Contains(exclude, k) {
			continue
		}

		if shell {
			fmt.Fprintf(&s, "export %v=%v\n", k, helpers.EscapeShell(v))
			continue
		}

		fmt.Fprintf(&s, "%v\n", e)
	}

	return s.String()
}

func (env *environment) Image() string {
	return os.Getenv(env.config.ImageKey)
}

func (env *environment) DebugPackages() string {
	return os.Getenv(env.config.DebugPackagesKey)
}

func (env *environment) Config() *EnvironmentConfig {
	return env.config
}

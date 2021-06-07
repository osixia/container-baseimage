package core

import (
	"fmt"
	"os"

	"github.com/osixia/container-baseimage/errors"
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

func (env *environment) Image() string {
	return os.Getenv(env.config.ImageKey)
}

func (env *environment) DebugPackages() string {
	return os.Getenv(env.config.DebugPackagesKey)
}

func (env *environment) Config() *EnvironmentConfig {
	return env.config
}

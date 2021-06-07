package core

import (
	"fmt"
	"os"
)

type EnvironmentConfig struct {
	ImageNameKey string
	ImageTagKey  string

	DebugPackagesKey string
}

func (ec *EnvironmentConfig) Validate() (bool, error) {

	if ec.ImageNameKey == "" {
		return false, fmt.Errorf("%v: %w", "ImageNameKey", ErrRequired)
	}

	if ec.ImageTagKey == "" {
		return false, fmt.Errorf("%v: %w", "ImageTagKey", ErrRequired)
	}

	if ec.DebugPackagesKey == "" {
		return false, fmt.Errorf("%v: %w", "DebugPackagesKey", ErrRequired)
	}

	return true, nil
}

type Environment interface {
	ImageName() string
	ImageTag() string

	DebugPackages() string

	Config() *EnvironmentConfig
}

type environment struct {
	config *EnvironmentConfig
}

func NewEnvironment(ec *EnvironmentConfig) (Environment, error) {

	if _, err := ec.Validate(); err != nil {
		return nil, err
	}

	return &environment{
		config: ec,
	}, nil
}

func (e *environment) ImageName() string {
	return os.Getenv(e.config.ImageNameKey)
}

func (e *environment) ImageTag() string {
	return os.Getenv(e.config.ImageTagKey)
}

func (e *environment) DebugPackages() string {
	return os.Getenv(e.config.DebugPackagesKey)
}

func (e *environment) Config() *EnvironmentConfig {
	return e.config
}

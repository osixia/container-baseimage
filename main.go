package main

import (
	"context"
	"os"

	"github.com/osixia/container-baseimage/cmd"
	"github.com/osixia/container-baseimage/config"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/helpers"
	"github.com/osixia/container-baseimage/log"
)

func main() {

	// set logger environment variables configuration
	helpers.Mustf(log.SetEnvironmentConfig(config.LogEnvironmentConfig), "Error initializing logger environment")

	core.CoreConfig = config.CoreConfig

	// execute cmd
	mainCtx := context.Background()
	if err := cmd.Run(mainCtx); err != nil {
		os.Exit(1)
	}

}

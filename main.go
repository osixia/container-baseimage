package main

import (
	"context"
	"os"

	"github.com/osixia/container-baseimage/cmd"
	"github.com/osixia/container-baseimage/config"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/log"
)

func main() {

	// set logger environment variables configuration
	if err := log.SetEnvironmentConfig(config.LogEnvironmentConfig); err != nil {
		log.Fatalf("Error initializing logger environment: %v", err.Error())
	}

	// init core
	if err := core.Init(config.CoreConfig); err != nil {
		log.Fatalf("Error initializing core: %v", err.Error())
	}

	// execute cmd
	mainCtx := context.Background()
	if err := cmd.Run(mainCtx); err != nil {
		os.Exit(1)
	}

}

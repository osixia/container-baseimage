package main

import (
	"context"
	"os"

	"github.com/osixia/container-baseimage/ci/cmd"
)

func main() {

	// execute cmd
	mainCtx := context.Background()
	if err := cmd.Run(mainCtx); err != nil {
		os.Exit(1)
	}

}

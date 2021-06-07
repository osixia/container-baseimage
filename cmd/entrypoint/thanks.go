package entrypoint

import (
	"fmt"

	"github.com/osixia/container-baseimage/config"
	"github.com/osixia/container-baseimage/helpers"
)

var thanksFunc = func() string {
	return fmt.Sprintf("%v\n\nThanks to all contributors ♥", config.Contributors)
}

var thanksCmd = helpers.NewPrintCmd("thanks", "List contributors", "t", thanksFunc)

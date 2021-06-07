package entrypoint

import (
	"fmt"

	"github.com/osixia/container-baseimage/config"
)

var thanksFunc = func() string {
	return fmt.Sprintf("%v\n\nThanks to all contributors ♥", config.BuildContributors)
}

var thanksCmd = newPrintCmd("thanks", "List container-baseimage contributors", "t", thanksFunc)

package processes

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/osixia/container-baseimage/config"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/helpers"
)

func TestUnmatchedTagDoesNotStartOrStopOtherProcesses(t *testing.T) {

	dir := t.TempDir()
	cfg := *config.CoreConfig
	cfg.FilesystemConfig = &core.FilesystemConfig{RootPath: filepath.Join(dir, "container"), RunRootPath: filepath.Join(dir, "run"), EnvironmentFilesPrefix: ".env"}
	core.CoreConfig = &cfg

	if err := core.Init(&cfg); err != nil {
		t.Fatal(err)
	}

	c := core.Instance()

	if err := c.Filesystem().Create(); err != nil {
		t.Fatal(err)
	}

	p, err := c.Processes().Get("web")
	if err != nil {
		t.Fatal(err)
	}

	if err := helpers.CreateMarker(p.WantedDownFile()); err != nil {
		t.Fatal(err)
	}

	startCmd.Run(startCmd, []string{"tag:missing"})

	if !p.IsWantedDown() {
		t.Fatal("unmatched tag started unrelated process")
	}

	if err := os.Remove(p.WantedDownFile()); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(p.PIDFile(), []byte("999999999"), 0644); err != nil {
		t.Fatal(err)
	}

	stopCmd.Run(stopCmd, []string{"tag:missing"})

	if p.IsWantedDown() {
		t.Fatal("unmatched tag stopped unrelated process")
	}
}

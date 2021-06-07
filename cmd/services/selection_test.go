package services

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/osixia/container-baseimage/config"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/helpers"
)

func TestUnmatchedTagDoesNotMutateAllServices(t *testing.T) {

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

	serviceDir := filepath.Join(c.Filesystem().Paths().Services, "web")

	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		t.Fatal(err)
	}

	marker := filepath.Join(dir, "unexpected-execution")

	for _, script := range []string{"install.sh", "download.sh"} {
		if err := os.WriteFile(filepath.Join(serviceDir, script), []byte("#!/bin/sh\ntouch "+helpers.EscapeShell(marker)+"\n"), 0755); err != nil {
			t.Fatal(err)
		}
	}

	s, err := c.Services().Get("web")

	if err != nil {
		t.Fatal(err)
	}

	downloadCmd.Run(downloadCmd, []string{"tag:missing"})
	installCmd.Run(installCmd, []string{"tag:missing"})
	linkCmd.Run(linkCmd, []string{"tag:missing"})

	if s.IsLinked() {
		t.Fatal("unmatched tag linked unrelated service")
	}

	if err := s.Link(); err != nil {
		t.Fatal(err)
	}

	unlinkCmd.Run(unlinkCmd, []string{"tag:missing"})

	if !s.IsLinked() {
		t.Fatal("unmatched tag unlinked unrelated service")
	}

	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("unrelated scripts executed: %v", err)
	}
}

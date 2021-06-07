package core

import (
	"os"
	"path/filepath"
	"testing"
)

// Test helpers
// =============================

// same values as config/config.go, which imports core
var (
	testServicesConfig = &ServicesConfig{
		TagsDir: ".tags",

		DefaultPriority:  500,
		PriorityFilename: ".priority",

		OptionalFilename: ".optional",

		StopSignalFilename: ".stop-signal",

		DownloadFilename:   "download.sh",
		DownloadedFilename: ".downloaded",

		InstallFilename:   "install.sh",
		InstalledFilename: ".installed",

		StartupFilename: "startup.sh",
		ProcessFilename: "process.sh",
		FinishFilename:  "finish.sh",

		LinkedFilename: ".entrypoint",
	}

	testProcessesConfig = &ProcessesConfig{
		PIDFileSuffix:        ".pid",
		WantedDownFileSuffix: ".down",
		StartingFileSuffix:   ".starting",
		TagsDir:              ".tags",
	}
)

type testCore struct {
	fs   Filesystem
	svcs Services
	prcs Processes
}

func newTestCore(t *testing.T) *testCore {

	t.Helper()

	dir := t.TempDir()
	fs, err := newFilesystem(&FilesystemConfig{
		RootPath:               filepath.Join(dir, "container"),
		RunRootPath:            filepath.Join(dir, "run", "container"),
		EnvironmentFilesPrefix: ".env",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := fs.Create(); err != nil {
		t.Fatal(err)
	}

	svcs, err := newServices(fs, testServicesConfig)
	if err != nil {
		t.Fatal(err)
	}
	prcs, err := newProcesses(fs, testProcessesConfig)
	if err != nil {
		t.Fatal(err)
	}

	return &testCore{fs: fs, svcs: svcs, prcs: prcs}
}

// addService creates a service directory with the given files (relative
// path: content). Scripts are made executable.
func (tc *testCore) addService(t *testing.T, name string, files map[string]string) string {

	t.Helper()

	dir := filepath.Join(tc.fs.Paths().Services, name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	for f, content := range files {
		path := filepath.Join(dir, f)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		perm := os.FileMode(0644)
		if filepath.Ext(f) == ".sh" {
			perm = 0755
		}
		if err := os.WriteFile(path, []byte(content), perm); err != nil {
			t.Fatal(err)
		}
	}

	return dir
}

// runFile writes a file in the processes runtime directory.
func (tc *testCore) runFile(t *testing.T, name string, content string) string {

	t.Helper()

	path := filepath.Join(tc.fs.Paths().RunProcesses, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	return path
}

func exists(path string) bool {

	_, err := os.Lstat(path)
	return err == nil
}

func processNames(ps []Process) []string {

	names := []string{}
	for _, p := range ps {
		names = append(names, p.Name())
	}
	return names
}

func serviceNames(ss []Service) []string {

	names := []string{}
	for _, s := range ss {
		names = append(names, s.Name())
	}
	return names
}

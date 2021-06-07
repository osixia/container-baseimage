package core

import (
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Process state
// =============================

func TestProcessIsUpReadsOnlyValidPIDFiles(t *testing.T) {

	tc := newTestCore(t)

	for content, up := range map[string]bool{
		"42":                     true,
		"42\n":                   true,
		"0":                      false,
		"-1":                     false,
		"abc":                    false,
		"":                       false,
		strings.Repeat("1", 100): false,
	} {
		tc.runFile(t, "web.pid", content)
		p, err := tc.prcs.Get("web")
		if err != nil {
			t.Fatal(err)
		}
		if p.IsUp() != up {
			t.Errorf("pid file %q: got up %v, want %v", content, p.IsUp(), up)
		}
	}
}

func TestProcessIsUpRefusesSymlinkedPIDFile(t *testing.T) {

	tc := newTestCore(t)
	target := filepath.Join(t.TempDir(), "target")
	if err := os.WriteFile(target, []byte("42"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(tc.fs.Paths().RunProcesses, "web.pid")); err != nil {
		t.Fatal(err)
	}

	p, err := tc.prcs.Get("web")
	if err != nil {
		t.Fatal(err)
	}
	if p.IsUp() {
		t.Fatal("a symlinked pid file must not be read")
	}
}

func TestProcessUpAndDownStates(t *testing.T) {

	tc := newTestCore(t)
	p, err := tc.prcs.Get("web")
	if err != nil {
		t.Fatal(err)
	}

	// stopping: stop requested, process still running
	tc.runFile(t, "web.pid", "42")
	tc.runFile(t, "web.down", "")
	if p.IsUp() || p.IsDown() {
		t.Fatalf("stopping process: got up %v, down %v", p.IsUp(), p.IsDown())
	}

	// down: stop requested, process exited
	if err := os.Remove(p.PIDFile()); err != nil {
		t.Fatal(err)
	}
	if p.IsUp() || !p.IsDown() {
		t.Fatalf("stopped process: got up %v, down %v", p.IsUp(), p.IsDown())
	}

	// starting: no stop request, no pid file yet
	if err := p.SetWantedDown(false); err != nil {
		t.Fatal(err)
	}
	if p.IsUp() || p.IsDown() {
		t.Fatalf("starting process: got up %v, down %v", p.IsUp(), p.IsDown())
	}
}

func TestProcessesRejectInvalidNames(t *testing.T) {

	tc := newTestCore(t)

	for _, name := range []string{"", ".", "..", "../web", "a/b"} {
		if _, err := tc.prcs.Get(name); err == nil {
			t.Errorf("%q: expected an error", name)
		}
	}
}

// Start and stop requests
// =============================

// fakeSupervisor reacts to stop requests like the lifecycle: it removes the
// pid file of a process wanted down and writes it back once started again.
func fakeSupervisor(t *testing.T, p Process, done <-chan struct{}) <-chan struct{} {

	t.Helper()
	stopped := make(chan struct{})

	go func() {
		defer close(stopped)
		for {
			select {
			case <-done:
				return
			case <-time.After(20 * time.Millisecond):
			}
			if p.IsWantedDown() {
				_ = os.Remove(p.PIDFile())
			} else if _, err := os.Lstat(p.PIDFile()); os.IsNotExist(err) {
				_ = os.WriteFile(p.PIDFile(), []byte(strconv.Itoa(os.Getpid())), 0644)
			}
		}
	}()
	return stopped
}

func TestProcessesStopThenStart(t *testing.T) {

	tc := newTestCore(t)
	tc.runFile(t, "web.pid", "42")
	p, err := tc.prcs.Get("web")
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	stopped := fakeSupervisor(t, p, done)
	defer func() {
		close(done)
		<-stopped
	}()

	if err := tc.prcs.Stop([]Process{p}, 10*time.Second); err != nil {
		t.Fatal(err)
	}
	if !p.IsDown() {
		t.Fatal("process must be down after stop")
	}

	// another command, e.g. run by another user, starts it again
	p2, err := tc.prcs.Get("web")
	if err != nil {
		t.Fatal(err)
	}
	if err := tc.prcs.Start([]Process{p2}, 10*time.Second); err != nil {
		t.Fatal(err)
	}
	if !p2.IsUp() || exists(p2.WantedDownFile()) {
		t.Fatal("process must be up without stop request after start")
	}
}

func TestProcessesStopNotRunning(t *testing.T) {

	tc := newTestCore(t)
	p, err := tc.prcs.Get("misspelled")
	if err != nil {
		t.Fatal(err)
	}

	if err := tc.prcs.Stop([]Process{p}, time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if exists(p.WantedDownFile()) {
		t.Fatal("no stop request must be left for a process not running")
	}
}

func TestProcessesStopTimeout(t *testing.T) {

	tc := newTestCore(t)
	tc.runFile(t, "web.pid", "42")
	p, err := tc.prcs.Get("web")
	if err != nil {
		t.Fatal(err)
	}

	// no supervisor: the pid file is never removed
	if err := tc.prcs.Stop([]Process{p}, 10*time.Millisecond); err == nil {
		t.Fatal("expected a timeout error")
	}
}

// List
// =============================

func TestProcessesList(t *testing.T) {

	tc := newTestCore(t)
	tc.runFile(t, "up.pid", "42")
	tc.runFile(t, "stopped.down", "")
	tc.runFile(t, "restarting.starting", "")
	tc.runFile(t, "unrelated", "")
	// any user can create files here: an invalid name must be ignored
	tc.runFile(t, `bad\name.pid`, "42")
	// a tagged process keeps its tags directory while restarting
	tc.runFile(t, "tagged/.tags/web", "")
	tc.runFile(t, "up/.tags/web", "")

	ps, err := tc.prcs.List()
	if err != nil {
		t.Fatal(err)
	}
	if got, want := processNames(ps), []string{"restarting", "stopped", "up"}; !slices.Equal(got, want) {
		t.Errorf("all: got %v, want %v", got, want)
	}

	ps, err = tc.prcs.List(WithProcessesWantedDown(true))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := processNames(ps), []string{"stopped"}; !slices.Equal(got, want) {
		t.Errorf("wanted down: got %v, want %v", got, want)
	}

	ps, err = tc.prcs.List(WithProcessesTags([]string{"web"}))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := processNames(ps), []string{"tagged", "up"}; !slices.Equal(got, want) {
		t.Errorf("tag web: got %v, want %v", got, want)
	}

	ps, err = tc.prcs.List(WithProcessesNames([]string{"stopped", "tag:web"}), HandleProcessesTagPrefixInNames("tag:"))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := processNames(ps), []string{"stopped", "tagged", "up"}; !slices.Equal(got, want) {
		t.Errorf("names and tags: got %v, want %v", got, want)
	}
}

func TestProcessesListExplicitEmptySelection(t *testing.T) {

	tc := newTestCore(t)
	tc.runFile(t, "up.pid", "42")

	for name, opt := range map[string]ProcessesListOption{
		"names": WithProcessesNames([]string{}),
		"tags":  WithProcessesTags([]string{}),
	} {
		ps, err := tc.prcs.List(opt)
		if err != nil {
			t.Fatal(err)
		}
		if len(ps) != 0 {
			t.Errorf("explicit empty %v must select nothing, got %v", name, processNames(ps))
		}
	}
}

func TestProcessesListRejectsInvalidName(t *testing.T) {

	tc := newTestCore(t)

	if _, err := tc.prcs.List(WithProcessesNames([]string{"../etc"})); err == nil {
		t.Fatal("expected an error")
	}
}

// New and cleanup
// =============================

func TestProcessesNewReplacesStaleTags(t *testing.T) {

	tc := newTestCore(t)
	tc.addService(t, "web", map[string]string{"process.sh": "#!/bin/sh\n", ".tags/http": ""})
	tc.runFile(t, "web/.tags/removed", "")
	s, err := tc.svcs.Get("web")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := tc.prcs.New(s); err != nil {
		t.Fatal(err)
	}

	p, err := tc.prcs.Get("web")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := p.Tags(), []string{"http"}; !slices.Equal(got, want) {
		t.Fatalf("got tags %v, want %v", got, want)
	}
}

func TestProcessesCleanupOnlyOwnServices(t *testing.T) {

	tc := newTestCore(t)
	tc.addService(t, "mine", map[string]string{"process.sh": "#!/bin/sh\n"})
	tc.addService(t, "hook-only", map[string]string{"startup.sh": "#!/bin/sh\n"})
	mine, err := tc.svcs.Get("mine")
	if err != nil {
		t.Fatal(err)
	}
	hookOnly, err := tc.svcs.Get("hook-only")
	if err != nil {
		t.Fatal(err)
	}

	var removed, kept []string
	for _, f := range []string{"mine.pid", "mine.down", "mine.starting"} {
		removed = append(removed, tc.runFile(t, f, "42"))
	}
	// files of services supervised by another container sharing the directory
	for _, f := range []string{"other.pid", "other.down", "other.starting", "hook-only.down"} {
		kept = append(kept, tc.runFile(t, f, "42"))
	}

	tc.prcs.CleanupRuntimeFiles([]Service{mine, hookOnly})

	for _, f := range removed {
		if exists(f) {
			t.Errorf("%v must be removed", f)
		}
	}
	for _, f := range kept {
		if !exists(f) {
			t.Errorf("%v must be kept", f)
		}
	}
}

func TestProcessesCleanupKeepsSymlinkTarget(t *testing.T) {

	tc := newTestCore(t)
	tc.addService(t, "mine", map[string]string{"process.sh": "#!/bin/sh\n"})
	mine, err := tc.svcs.Get("mine")
	if err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(t.TempDir(), "target")
	if err := os.WriteFile(target, []byte("outside"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(tc.fs.Paths().RunProcesses, "mine.pid")); err != nil {
		t.Fatal(err)
	}

	tc.prcs.CleanupRuntimeFiles([]Service{mine})

	data, err := os.ReadFile(target)
	if err != nil || string(data) != "outside" {
		t.Fatalf("symlink target must stay intact, got %q, %v", data, err)
	}
}

package core

import (
	"context"
	goerrors "errors"
	"os"
	"path/filepath"
	"slices"
	"syscall"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watcher reconciliation
// =============================

// The channels let tests deliver stale events and overflows deterministically,
// while process state is read from the real runtime files.
func runTestProcessWatcher(t *testing.T, lc *lifecycle) *fsnotify.Watcher {

	t.Helper()
	w := &fsnotify.Watcher{
		Events: make(chan fsnotify.Event),
		Errors: make(chan error),
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		lc.watchProcesses(w)
	}()
	t.Cleanup(func() {
		close(w.Events)
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Error("process watcher did not stop after closing its events")
		}
	})
	return w
}

func sendProcessEvent(t *testing.T, w *fsnotify.Watcher, event fsnotify.Event) {

	t.Helper()
	select {
	case w.Events <- event:
	case <-time.After(5 * time.Second):
		t.Fatal("process watcher did not receive an event")
	}
}

func TestLifecycleWatcherReadsCurrentStopState(t *testing.T) {
	for name, tt := range map[string]struct {
		op         fsnotify.Op
		wantedDown bool
	}{
		"stop request":        {fsnotify.Create, true},
		"stale stop request":  {fsnotify.Create, false},
		"stale start request": {fsnotify.Remove, true},
	} {
		t.Run(name, func(t *testing.T) {
			tc := newTestCore(t)
			p, err := tc.prcs.Get("web")
			if err != nil {
				t.Fatal(err)
			}
			// A queued event can refer to an earlier generation of the process.
			previous, cancelPrevious := context.WithCancel(context.Background())
			defer cancelPrevious()
			current, cancelCurrent := context.WithCancel(context.Background())
			defer cancelCurrent()
			lcp := &lifecycleProcess{process: p, cancel: cancelPrevious}
			lcp.mu.Lock()
			lcp.cancel = cancelCurrent
			lcp.mu.Unlock()
			if err := p.SetWantedDown(tt.wantedDown); err != nil {
				t.Fatal(err)
			}

			w := runTestProcessWatcher(t, &lifecycle{processes: []*lifecycleProcess{lcp}})
			sendProcessEvent(t, w, fsnotify.Event{Name: p.WantedDownFile(), Op: tt.op})
			// Receiving a second event means the previous one was fully handled.
			sendProcessEvent(t, w, fsnotify.Event{Name: filepath.Join(tc.fs.Paths().RunProcesses, "unrelated")})

			if got := current.Err() != nil; got != tt.wantedDown {
				t.Fatalf("current process cancelled: %v, want %v", got, tt.wantedDown)
			}
			if previous.Err() != nil {
				t.Fatal("a stale event must not cancel the previous process generation")
			}
		})
	}
}

func TestLifecycleWatcherReconcilesAfterError(t *testing.T) {

	tc := newTestCore(t)
	contexts := make(map[string]context.Context)
	lc := &lifecycle{}
	for _, name := range []string{"stopped", "running"} {
		p, err := tc.prcs.Get(name)
		if err != nil {
			t.Fatal(err)
		}
		if err := p.SetWantedDown(name == "stopped"); err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)
		contexts[name] = ctx
		lc.processes = append(lc.processes, &lifecycleProcess{process: p, cancel: cancel})
	}
	w := runTestProcessWatcher(t, lc)
	select {
	case w.Errors <- goerrors.New("event queue overflow"):
	case <-time.After(5 * time.Second):
		t.Fatal("process watcher did not receive an error")
	}
	sendProcessEvent(t, w, fsnotify.Event{Name: "unrelated"})
	if contexts["stopped"].Err() == nil {
		t.Fatal("the lost stop event must be recovered after a watcher error")
	}
	if contexts["running"].Err() != nil {
		t.Fatal("reconciliation must preserve a process still wanted up")
	}
}

// Signal forwarding
// =============================

type recordingSignaler struct {
	signals []os.Signal
	err     error
}

func (s *recordingSignaler) Signal(sig os.Signal) error {

	s.signals = append(s.signals, sig)
	return s.err
}

func TestLifecycleForwardSignalsFiltersSupervisorSignals(t *testing.T) {

	signals := []os.Signal{
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGCHLD, syscall.SIGURG,
		syscall.SIGHUP, syscall.SIGUSR1, syscall.SIGUSR2,
	}
	ch := make(chan os.Signal, len(signals))
	for _, sig := range signals {
		ch <- sig
	}
	close(ch)
	finished := &recordingSignaler{err: os.ErrProcessDone}
	running := &recordingSignaler{}
	ForwardSignals(ch, finished, running)
	want := []os.Signal{syscall.SIGHUP, syscall.SIGUSR1, syscall.SIGUSR2}
	for _, s := range []*recordingSignaler{finished, running} {
		if !slices.Equal(s.signals, want) {
			t.Errorf("forwarded signals: %v, want %v", s.signals, want)
		}
	}
}

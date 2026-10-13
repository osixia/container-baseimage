package helpers

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

func TestReaperPreservesManagedExitStatus(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() {
		defer close(done)

		for ctx.Err() == nil {
			reapChildren()
			time.Sleep(time.Millisecond)
		}
	}()

	var wg sync.WaitGroup

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := NewExec(ctx).WithNoLog(true).Command("sh", "-c", "exit 7").Run()

			var exit *exec.ExitError

			if !errors.As(err, &exit) || exit.ExitCode() != 7 {
				t.Errorf("lost exit status: %v", err)
			}
		}()
	}

	wg.Wait()
	cancel()
	<-done
}

func TestReaperCollectsUnmanagedChild(t *testing.T) {

	cmd := exec.Command("sh", "-c", "exit 0")

	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	defer func() { _ = cmd.Process.Release() }()
	deadline := time.Now().Add(2 * time.Second)

	for time.Now().Before(deadline) {
		reapChildren()

		if err := syscall.Kill(cmd.Process.Pid, 0); err == syscall.ESRCH {
			return
		}

		time.Sleep(time.Millisecond)
	}

	_ = cmd.Wait()
	t.Fatal("unmanaged child was not reaped")
}

func TestExecConcurrentSignalsAndCancellation(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	e := NewExec(ctx).WithSetGPID(true).WithTimeout(100*time.Millisecond).WithNoLog(true).Command("sh", "-c", "trap '' INT TERM; while :; do sleep 1; done")

	if err := e.Start(); err != nil {
		t.Fatal(err)
	}

	done := make(chan error, 1)
	go func() {
		done <- e.Wait()
	}()

	var wg sync.WaitGroup

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for j := 0; j < 50; j++ {
				_ = e.Signal(syscall.SIGCONT)
			}
		}()
	}

	cancel()
	wg.Wait()

	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}

	case <-time.After(2 * time.Second):
		_ = e.Signal(os.Kill)
		t.Fatal("process did not stop")
	}

	if !errors.Is(e.Signal(syscall.SIGCONT), os.ErrProcessDone) {
		t.Fatal("finished process remains signalable")
	}
}

func TestExecPreservesFailureWhenContextCancelledAfterExit(t *testing.T) {

	for _, timeout := range []time.Duration{0, 100 * time.Millisecond} {
		t.Run(timeout.String(), func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			e := NewExec(ctx).WithSetGPID(true).WithTimeout(timeout).WithNoLog(true).Command("sh", "-c", "exit 7")
			if err := e.Start(); err != nil {
				t.Fatal(err)
			}

			// Observe the zombie in /proc without consuming its exit status.
			deadline := time.Now().Add(2 * time.Second)
			exited := false
			for time.Now().Before(deadline) {
				data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", e.process.Pid))
				if err != nil {
					t.Fatal(err)
				}

				end := strings.LastIndexByte(string(data), ')')
				if end >= 0 && strings.HasPrefix(string(data[end+1:]), " Z ") {
					exited = true
					break
				}

				time.Sleep(time.Millisecond)
			}

			if !exited {
				_ = e.Signal(os.Kill)
				_ = e.Wait()
				t.Fatal("child did not exit before cancellation")
			}

			cancel()
			err := e.Wait()
			var exitErr *exec.ExitError
			if !errors.As(err, &exitErr) || exitErr.ExitCode() != 7 {
				t.Fatalf("exit failure was hidden: %v", err)
			}
		})
	}
}

func TestExecDrainsLongOutputLines(t *testing.T) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var stdout, stderr []string
	e := NewExec(ctx).WithTimeout(100*time.Millisecond).WithNoLog(true).
		WithStdoutLogPrintFunc(func(line string) { stdout = append(stdout, line) }).
		WithStderrLogPrintFunc(func(line string) { stderr = append(stderr, line) }).
		Command("sh", "-c", "head -c 1048576 /dev/zero | tr '\\000' x; printf '\\nlast'; head -c 1048576 /dev/zero | tr '\\000' y >&2; printf '\\nend' >&2")

	if err := e.Run(); err != nil {
		t.Fatal(err)
	}

	for name, lines := range map[string][]string{"stdout": stdout, "stderr": stderr} {
		if len(lines) != 2 || len(lines[0]) != 1048576 {
			t.Fatalf("%s output was truncated: %d lines", name, len(lines))
		}
	}

	if stdout[1] != "last" || stderr[1] != "end" {
		t.Fatalf("unterminated lines lost: stdout=%q stderr=%q", stdout[1], stderr[1])
	}
}

func TestExecCancellationKillsProcessGroup(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	marker := filepath.Join(t.TempDir(), "child")
	script := "trap '' INT TERM\nsh -c 'trap \"\" INT TERM; exec sleep 30' &\necho $! > " + EscapeShell(marker) + "\nwait\n"
	e := NewExec(ctx).WithSetGPID(true).WithTimeout(50*time.Millisecond).WithNoLog(true).Command("sh", "-c", script)
	if err := e.Start(); err != nil {
		t.Fatal(err)
	}

	deadline := time.Now().Add(2 * time.Second)
	var data []byte
	for time.Now().Before(deadline) {
		var err error
		data, err = os.ReadFile(marker)
		if err == nil && len(data) > 0 {
			break
		}

		time.Sleep(time.Millisecond)
	}

	child, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		_ = e.Signal(os.Kill)
		_ = e.Wait()
		t.Fatal("child did not become ready", err)
	}

	defer func() { _ = syscall.Kill(child, syscall.SIGKILL) }()
	cancel()
	if err := e.Wait(); err != nil {
		t.Fatal(err)
	}

	deadline = time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", child))
		if errors.Is(err, os.ErrNotExist) {
			return
		}

		end := strings.LastIndexByte(string(data), ')')
		if end >= 0 && strings.HasPrefix(string(data[end+1:]), " Z ") {
			return
		}

		time.Sleep(time.Millisecond)
	}

	t.Fatal("child survived process-group cancellation")
}

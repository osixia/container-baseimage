package helpers

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/osixia/container-baseimage/errors"
)

// Test helpers
// =============================

// shellExec runs script with sh, $D being dir. Outputs are discarded: sh
// reports background jobs killed by the tests.
func shellExec(ctx context.Context, dir string, script string) *Exec {

	discard := func(string) {}
	return NewExec(ctx).WithNoLog(true).WithStdoutLogPrintFunc(discard).WithStderrLogPrintFunc(discard).
		WithEnv(append(os.Environ(), "D="+dir)).Command("/bin/sh", "-c", script)
}

func waitForFile(t *testing.T, path string) {

	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for {
		if _, err := os.Stat(path); err == nil {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %v", path)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func readPIDFile(t *testing.T, path string) int {

	t.Helper()
	waitForFile(t, path)
	deadline := time.Now().Add(10 * time.Second)
	for {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if pid, err := strconv.Atoi(strings.TrimSpace(string(data))); err == nil && pid > 0 {
			return pid
		}
		if time.Now().After(deadline) {
			t.Fatalf("invalid pid file %v: %q", path, data)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// processGone reports an exited process, zombies included: the reparented
// descendants are reaped by another process, possibly later.
func processGone(pid int) bool {

	if err := syscall.Kill(pid, 0); err == syscall.ESRCH {
		return true
	}
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return true
	}
	end := bytes.LastIndexByte(data, ')')
	fields := strings.Fields(string(data[end+1:]))
	return len(fields) > 0 && (fields[0] == "Z" || fields[0] == "X")
}

func waitProcessGone(t *testing.T, pid int, timeout time.Duration) {

	t.Helper()
	deadline := time.Now().Add(timeout)
	for !processGone(pid) {
		if time.Now().After(deadline) {
			_ = syscall.Kill(pid, syscall.SIGKILL)
			t.Fatalf("process %v still running", pid)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// Exit codes
// =============================

func TestExecPreservesExitCode(t *testing.T) {

	err := shellExec(context.Background(), t.TempDir(), "exit 3").Run()

	if code := errors.ExitCode(err); code != 3 {
		t.Fatalf("got exit code %v (%v), want 3", code, err)
	}
}

func TestExecExitCodeOfSignaledProcess(t *testing.T) {

	err := shellExec(context.Background(), t.TempDir(), "kill -KILL $$").Run()

	if code := errors.ExitCode(err); code != 128+int(syscall.SIGKILL) {
		t.Fatalf("got exit code %v (%v), want %v", code, err, 128+int(syscall.SIGKILL))
	}
}

func TestExecKeepsExitErrorDuringCancellation(t *testing.T) {

	dir := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	e := shellExec(ctx, dir, `trap 'exit 7' TERM; touch "$D/ready"; while :; do sleep 0.05; done`).
		WithSetGPID(true).WithTimeout(5 * time.Second)
	if err := e.Start(); err != nil {
		t.Fatal(err)
	}
	waitForFile(t, filepath.Join(dir, "ready"))

	cancel()
	err := e.Wait()

	if code := errors.ExitCode(err); code != 7 {
		t.Fatalf("a real exit failure must survive cancellation, got %v (%v)", code, err)
	}
}

// Process groups and signals
// =============================

func TestExecCancellationStopsProcessGroup(t *testing.T) {

	dir := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	e := shellExec(ctx, dir, `sleep 60 & echo $! > "$D/child.pid"; wait`).
		WithSetGPID(true).WithTimeout(5 * time.Second)
	if err := e.Start(); err != nil {
		t.Fatal(err)
	}
	child := readPIDFile(t, filepath.Join(dir, "child.pid"))

	cancel()
	if err := e.Wait(); err != nil {
		t.Fatalf("a cancelled process stopped by its stop signal is not an error: %v", err)
	}

	waitProcessGone(t, child, 5*time.Second)
}

func TestExecCancellationEscalatesToSIGKILL(t *testing.T) {

	dir := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	const stopDelay = 200 * time.Millisecond

	// an ignored signal stays ignored in the descendants
	e := shellExec(ctx, dir, `trap '' TERM; sleep 60 & echo $! > "$D/child.pid"; while :; do sleep 0.05; done`).
		WithSetGPID(true).WithTimeout(stopDelay)
	if err := e.Start(); err != nil {
		t.Fatal(err)
	}
	child := readPIDFile(t, filepath.Join(dir, "child.pid"))

	start := time.Now()
	cancel()
	err := e.Wait()

	if elapsed := time.Since(start); elapsed < stopDelay || elapsed > 5*time.Second {
		t.Fatalf("stop must respect the grace period and remain bounded, took %v (delay %v)", elapsed, stopDelay)
	}
	if err != nil {
		t.Fatalf("a cancelled process killed after the stop delay is not an error: %v", err)
	}
	waitProcessGone(t, child, 5*time.Second)
}

func TestExecStopSignalOverride(t *testing.T) {

	dir := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	e := shellExec(ctx, dir, `trap 'touch "$D/quit"; exit 0' QUIT; trap '' TERM; touch "$D/ready"; while :; do sleep 0.05; done`).
		WithSetGPID(true).WithTimeout(5 * time.Second).WithStopSignal(syscall.SIGQUIT)
	if err := e.Start(); err != nil {
		t.Fatal(err)
	}
	waitForFile(t, filepath.Join(dir, "ready"))

	cancel()
	if err := e.Wait(); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, "quit")); err != nil {
		t.Fatalf("the configured stop signal must be sent: %v", err)
	}
}

func TestExecSignalOnlyReachesMainProcess(t *testing.T) {

	dir := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	script := `trap 'touch "$D/leader.hup"' HUP
sh -c 'trap "touch \"\$D/child.hup\"" HUP; touch "$D/child.ready"; while :; do sleep 0.05; done' &
touch "$D/leader.ready"
while :; do sleep 0.05; done`
	e := shellExec(ctx, dir, script).WithSetGPID(true).WithTimeout(5 * time.Second)
	if err := e.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		cancel()
		_ = e.Wait()
	}()
	waitForFile(t, filepath.Join(dir, "leader.ready"))
	waitForFile(t, filepath.Join(dir, "child.ready"))

	if err := e.Signal(syscall.SIGHUP); err != nil {
		t.Fatal(err)
	}

	waitForFile(t, filepath.Join(dir, "leader.hup"))
	time.Sleep(300 * time.Millisecond)
	if _, err := os.Stat(filepath.Join(dir, "child.hup")); err == nil {
		t.Fatal("relayed signals must not reach the process group (e.g. keepalived VRRP child)")
	}
}

func TestExecStopGroupOnExit(t *testing.T) {

	dir := t.TempDir()

	// the child must not hold the logged outputs pipes: Wait would wait for
	// them until WaitDelay (services write to the supervisor outputs directly)
	e := shellExec(context.Background(), dir, `sleep 60 >/dev/null 2>&1 & echo $! > "$D/child.pid"`).
		WithSetGPID(true).WithStopGroupOnExit(true).WithTimeout(2 * time.Second)
	if err := e.Run(); err != nil {
		t.Fatal(err)
	}

	child := readPIDFile(t, filepath.Join(dir, "child.pid"))
	waitProcessGone(t, child, 5*time.Second)
}

func TestExecSignalAfterExit(t *testing.T) {

	e := shellExec(context.Background(), t.TempDir(), "exit 0")
	if err := e.Run(); err != nil {
		t.Fatal(err)
	}

	if err := e.Signal(syscall.SIGHUP); err != os.ErrProcessDone {
		t.Fatalf("got %v, want %v", err, os.ErrProcessDone)
	}
}

// PID file
// =============================

func TestExecPIDFileLifecycle(t *testing.T) {

	dir := t.TempDir()
	pidFile := filepath.Join(dir, "service.pid")

	e := shellExec(context.Background(), dir, `while [ ! -e "$D/go" ]; do sleep 0.05; done`).WithPIDFile(pidFile)
	if err := e.Start(); err != nil {
		t.Fatal(err)
	}

	if pid := readPIDFile(t, pidFile); pid != e.Cmd.Process.Pid {
		t.Fatalf("pid file holds %v, want %v", pid, e.Cmd.Process.Pid)
	}

	writeTestFile(t, filepath.Join(dir, "go"), "")
	if err := e.Wait(); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Lstat(pidFile); !os.IsNotExist(err) {
		t.Fatalf("pid file must be removed once the process exited: %v", err)
	}
}

// Logged command
// =============================

func TestExecErrorDoesNotShowArguments(t *testing.T) {

	err := shellExec(context.Background(), t.TempDir(), "exit 1 # password=secret").Run()

	if err == nil || strings.Contains(err.Error(), "secret") {
		t.Fatalf("command arguments may hold secrets and must not appear in errors: %v", err)
	}
}

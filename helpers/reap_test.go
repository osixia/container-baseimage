package helpers

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/osixia/container-baseimage/errors"
)

func TestReadPPIDWithParenthesesInCommandName(t *testing.T) {

	sleep, err := exec.LookPath("sleep")
	if err != nil {
		t.Skip("sleep not found")
	}
	if sleep, err = filepath.EvalSymlinks(sleep); err != nil {
		t.Fatal(err)
	}
	if filepath.Base(sleep) == "busybox" {
		t.Skip("busybox applets depend on their name")
	}

	// the command name field of /proc/<pid>/stat is not escaped
	name := filepath.Join(t.TempDir(), "a) 1 2 (b")
	copyExecutable(t, sleep, name)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	e := NewExec(ctx).WithNoLog(true).Command(name, "60")
	if err := e.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		cancel()
		_ = e.Wait()
	}()

	ppid, err := readPPID(e.Cmd.Process.Pid)
	if err != nil {
		t.Fatal(err)
	}
	if ppid != os.Getpid() {
		t.Fatalf("got ppid %v, want %v", ppid, os.Getpid())
	}
}

func TestReapChildrenKeepsManagedExitStatus(t *testing.T) {

	e := shellExec(context.Background(), t.TempDir(), "exit 5")
	if err := e.Start(); err != nil {
		t.Fatal(err)
	}

	// the reaper runs while the managed child is a zombie, before Wait
	pid := e.Cmd.Process.Pid
	deadline := time.Now().Add(10 * time.Second)
	for !processGone(pid) {
		if time.Now().After(deadline) {
			t.Fatal("child did not exit")
		}
		time.Sleep(10 * time.Millisecond)
	}
	reapChildren()

	if code := errors.ExitCode(e.Wait()); code != 5 {
		t.Fatalf("the reaper must not steal a managed child status, got exit code %v", code)
	}
}

func TestReapChildrenReapsUnmanagedChild(t *testing.T) {

	pid, err := syscall.ForkExec("/bin/sh", []string{"sh", "-c", "exit 0"}, &syscall.ProcAttr{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _, _ = syscall.Wait4(pid, nil, 0, nil) }()

	deadline := time.Now().Add(10 * time.Second)
	for !processGone(pid) {
		if time.Now().After(deadline) {
			t.Fatal("child did not exit")
		}
		time.Sleep(10 * time.Millisecond)
	}

	reapChildren()

	if _, err := syscall.Wait4(pid, nil, syscall.WNOHANG, nil); err != syscall.ECHILD {
		t.Fatalf("unmanaged zombie must be reaped, wait4 returned %v", err)
	}
}

func copyExecutable(t *testing.T, src string, dst string) {

	t.Helper()
	in, err := os.Open(src)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = in.Close() }()
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0755)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.Copy(out, in); err != nil {
		t.Fatal(err)
	}
	if err := out.Close(); err != nil {
		t.Fatal(err)
	}
}

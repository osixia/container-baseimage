package helpers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/osixia/container-baseimage/log"
)

// Register managed children before a reaper can see them. Never wait on a PID
// owned by exec.Cmd: its Wait call must retain the exit status.
var childProcesses = struct {
	sync.Mutex
	managed map[int]bool
}{managed: make(map[int]bool)}

func ReapChildren(ctx context.Context) {

	log.Trace("ReapChildren called")

	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			reapChildren()
		}
	}
}

// reapChildren discovers our direct, unmanaged children by scanning /proc and
// reading each candidate's ppid from /proc/<pid>/stat rather than relying on
// /proc/<pid>/task/<tid>/children, which requires the kernel to be built with
// CONFIG_CHECKPOINT_RESTORE and is not implemented by every container runtime
// sandbox (for example gVisor/runsc). /proc/<pid>/stat is basic, universally
// supported procfs and is sufficient since a child's reported ppid is always
// its parent *process* id, regardless of which OS thread called fork.
func reapChildren() {

	log.Trace("reapChildren called")

	childProcesses.Lock()
	defer childProcesses.Unlock()

	entries, err := os.ReadDir("/proc")

	if err != nil {
		log.Warningf("Failed to list processes: %v", err)
		return
	}

	self := os.Getpid()

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pid, err := strconv.Atoi(entry.Name())
		if err != nil || pid <= 0 {
			continue
		}

		if childProcesses.managed[pid] {
			continue
		}

		ppid, err := readPPID(pid)
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				log.Debugf("Failed to read parent pid for %v: %v", pid, err)
			}
			continue
		}

		if ppid != self {
			continue
		}

		var status syscall.WaitStatus

		// Wait for this specific unmanaged child, never for an arbitrary child.
		var reaped int
		for {
			reaped, err = syscall.Wait4(pid, &status, syscall.WNOHANG, nil)
			if !errors.Is(err, syscall.EINTR) {
				break
			}
		}

		if err != nil && !errors.Is(err, syscall.ECHILD) && !errors.Is(err, syscall.ESRCH) {
			log.Warningf("Failed to reap child PID %v: %v", pid, err)
		} else if reaped > 0 {
			log.Debugf("Reaped pid: %v", reaped)
		}
	}
}

// readPPID reads the parent process id of pid from /proc/<pid>/stat. The
// command name field is parenthesized and may itself contain spaces or
// parentheses, so the remaining fields are read after the last ')'.
func readPPID(pid int) (int, error) {

	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return 0, err
	}

	end := bytes.LastIndexByte(data, ')')
	if end < 0 || end+1 >= len(data) {
		return 0, fmt.Errorf("unexpected /proc/%v/stat format", pid)
	}

	// Fields after the command name: state ppid pgrp session ...
	fields := strings.Fields(string(data[end+1:]))
	if len(fields) < 2 {
		return 0, fmt.Errorf("unexpected /proc/%v/stat format", pid)
	}

	return strconv.Atoi(fields[1])
}

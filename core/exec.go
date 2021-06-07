package core

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/google/shlex"

	"github.com/osixia/container-baseimage/log"
)

type Exec struct {
	context context.Context
	timeout time.Duration

	stdout *bytes.Buffer
}

func NewExec(ctx context.Context) *Exec {

	return &Exec{
		context: ctx,
	}
}

func (e *Exec) WithTimeout(timeout time.Duration) *Exec {

	e.timeout = timeout

	return e
}

func (e *Exec) WithStdout(b *bytes.Buffer) *Exec {

	e.stdout = b

	return e
}

func (e *Exec) Command(args ...string) error {

	log.Infof("Running command %v ...", strings.Join(args, " "))

	return e.exec(args...)
}

func (e *Exec) Script(script string) error {

	log.Infof("Running script %v ...", script)

	return e.Shlex(script)
}

func (e *Exec) Scripts(scripts []string) error {

	for _, script := range scripts {
		if err := e.Script(script); err != nil {
			return err
		}
	}

	return nil
}

func (e *Exec) Shlex(args string) error {

	log.Tracef("Exec.Shlex called with args: %v", args)

	shCmd, err := shlex.Split(args)
	if err != nil {
		return err
	}

	return e.exec(shCmd...)
}

func (e *Exec) exec(cmds ...string) error {

	log.Tracef("Exec.exec called with cmds: %v", cmds)

	name, args := cmds[0], cmds[1:]

	cmd := exec.CommandContext(e.context, name, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr

	if e.stdout != nil {
		cmd.Stdout = e.stdout
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	log.Debugf("Command %v started with PID %v", strings.Join(cmds, " "), cmd.Process.Pid)

	if e.timeout > 0 {
		return e.waitOrStop(e.context, cmd, os.Interrupt, e.timeout)
	}

	return cmd.Wait()
}

// waitOrStop waits for the already-started command cmd by calling its Wait method.
//
// If cmd does not return before ctx is done, waitOrStop sends it the given interrupt signal.
// If killDelay is positive, waitOrStop waits that additional period for Wait to return before sending os.Kill.
//
// This function is copied from the one added to x/playground/internal in
// http://golang.org/cl/228438.
func (e *Exec) waitOrStop(ctx context.Context, cmd *exec.Cmd, interrupt os.Signal, killDelay time.Duration) error {

	log.Tracef("Exec.waitOrStop called with cmd: %v, interrupt: %v, killDelay: %v", cmd, interrupt, killDelay)

	if cmd.Process == nil {
		log.Fatal("waitOrStop called with a nil cmd.Process — missing Start call?")
	}
	if interrupt == nil {
		log.Fatal("waitOrStop requires a non-nil interrupt signal")
	}

	errc := make(chan error)
	go func() {
		select {
		case errc <- nil:
			return
		case <-ctx.Done():
		}

		err := cmd.Process.Signal(interrupt)
		if err == nil {
			err = ctx.Err() // Report ctx.Err() as the reason we interrupted.
		} else if err.Error() == "os: process already finished" {
			errc <- nil
			return
		}

		if killDelay > 0 {
			timer := time.NewTimer(killDelay)
			select {
			// Report ctx.Err() as the reason we interrupted the process...
			case errc <- ctx.Err():
				timer.Stop()
				return
			// ...but after killDelay has elapsed, fall back to a stronger signal.
			case <-timer.C:
			}

			// Wait still hasn't returned.
			// Kill the process harder to make sure that it exits.
			//
			// Ignore any error: if cmd.Process has already terminated, we still
			// want to send ctx.Err() (or the error from the Interrupt call)
			// to properly attribute the signal that may have terminated it.
			_ = cmd.Process.Kill()
		}

		errc <- err
	}()

	waitErr := cmd.Wait()
	if interruptErr := <-errc; interruptErr != nil {
		return interruptErr
	}
	return waitErr
}

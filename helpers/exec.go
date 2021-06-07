package helpers

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"

	"github.com/osixia/container-baseimage/log"
)

// Exec wrapper
// =============================

type Exec struct {
	context context.Context
	timeout time.Duration
	setGPID bool

	Cmd     *exec.Cmd
	pidFile string

	env []string

	stdoutLogPrintFunc log.PrintFunc
	stderrLogPrintFunc log.PrintFunc

	noLog bool
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

func (e *Exec) WithSetGPID(b bool) *Exec {
	e.setGPID = b

	return e
}

func (e *Exec) WithPIDFile(f string) *Exec {
	e.pidFile = f

	return e
}

func (e *Exec) WithEnv(env []string) *Exec {
	e.env = env

	return e
}

func (e *Exec) WithStdoutLogPrintFunc(fn log.PrintFunc) *Exec {
	e.stdoutLogPrintFunc = fn

	return e
}

func (e *Exec) WithStderrLogPrintFunc(fn log.PrintFunc) *Exec {
	e.stderrLogPrintFunc = fn

	return e
}

func (e *Exec) WithNoLog(b bool) *Exec {
	e.noLog = b

	return e
}

func (e *Exec) Command(name string, args ...string) *Exec {
	log.Tracef("Exec.Command called with cmd: %v, args: %v", name, args)

	e.Cmd = exec.CommandContext(e.context, name, args...)

	return e
}

func (e *Exec) Run() error {

	log.Trace("Exec.Run called")

	if err := e.Start(); err != nil {
		return err
	}

	return e.Wait()
}

func (e *Exec) Start() error {

	log.Trace("Exec.Start called")

	if !e.noLog {
		log.Infof("Running %v ...", e.Cmd)
	}

	if err := e.prepare(); err != nil {
		return err
	}

	if err := e.Cmd.Start(); err != nil {
		return err
	}

	pid := e.Cmd.Process.Pid

	if e.setGPID {
		e.Cmd.Cancel = func() error {
			return syscall.Kill(-pid, syscall.SIGTERM)
		}
	}

	if e.pidFile != "" {
		f, err := Create(e.pidFile)
		if err != nil {
			log.Warningf("Error creating pid file %v: %v", e.pidFile, err.Error())
		} else {
			defer func() {
				if err := f.Close(); err != nil {
					log.Error(err.Error())
				}
			}()

			if _, err := f.WriteString(strconv.Itoa(pid)); err != nil {
				log.Warningf("Error writing in pid file %v: %v", e.pidFile, err.Error())
			}
		}
	}

	log.Debugf("%v: started (pid %v)", e.Cmd, pid)

	return nil
}

func (e *Exec) Wait() error {

	log.Trace("Exec.Wait called")

	var err error
	if e.timeout > 0 {
		err = e.waitOrStop(e.context, e.Cmd, os.Interrupt, e.timeout)
	} else {
		err = e.Cmd.Wait()
	}

	if e.pidFile != "" {
		if err := Remove(e.pidFile); err != nil {
			log.Warningf("Error removing pid file %v: %v", e.pidFile, err.Error())
		}
	}

	if errors.Is(err, context.Canceled) {
		err = nil
	}

	if err != nil {
		err = fmt.Errorf("%v: %w", e.Cmd, err)
	}

	return err
}

// Signal sends sig to the process. If setGPID is true, the signal is sent to
// the entire process group so that child processes also receive it.
func (e *Exec) Signal(sig os.Signal) error {

	log.Tracef("Exec.Signal called with sig: %v", sig)

	if e.Cmd == nil || e.Cmd.Process == nil {
		return nil
	}

	var err error

	if e.setGPID {

		s, ok := sig.(syscall.Signal)
		if !ok {
			return fmt.Errorf("unsupported signal type %T", sig)
		}

		err = syscall.Kill(-e.Cmd.Process.Pid, s)
	} else {
		err = e.Cmd.Process.Signal(sig)
	}

	if err == nil {
		return nil
	}

	// Normalize "process already gone" errors.
	if errors.Is(err, os.ErrProcessDone) || errors.Is(err, syscall.ESRCH) {
		return os.ErrProcessDone
	}

	return err
}

func (e *Exec) prepare() error {

	log.Trace("Exec.prepare called")

	// set env
	if len(e.env) > 0 {
		e.Cmd.Env = e.env
	}

	// set inputs / outputs
	e.Cmd.Stdin = os.Stdin

	if err := e.setStd(&e.Cmd.Stdout, os.Stdout, e.stdoutLogPrintFunc, e.Cmd.StdoutPipe); err != nil {
		return err
	}

	if err := e.setStd(&e.Cmd.Stderr, os.Stderr, e.stderrLogPrintFunc, e.Cmd.StderrPipe); err != nil {
		return err
	}

	// set group id
	if e.setGPID {
		e.Cmd.SysProcAttr = &syscall.SysProcAttr{
			Setpgid: true,
		}
	}

	return nil
}

func (e *Exec) setStd(cmdStd *io.Writer, std io.Writer, printFunc log.PrintFunc, pipeFunc func() (io.ReadCloser, error)) error {

	log.Tracef("Exec.setStd called with cmdStd: %v, std: %v", cmdStd, std)

	if printFunc == nil {
		*cmdStd = std
		return nil
	}

	stdpipe, err := pipeFunc()
	if err != nil {
		return err
	}

	go e.printOutput(stdpipe, printFunc)

	return nil
}

func (e *Exec) printOutput(std io.Reader, fn log.PrintFunc) {

	log.Tracef("Exec.printOutput called with std: %v, fn: %v", std, fn)

	scanner := bufio.NewScanner(std)
	for scanner.Scan() {
		fn(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Tracef("failed to read std: %v", err)
	}
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

		err := e.Signal(interrupt)
		if err == nil {
			err = ctx.Err() // Report ctx.Err() as the reason we interrupted.
		} else if errors.Is(err, os.ErrProcessDone) {
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
			_ = e.Signal(os.Kill)
		}

		errc <- err
	}()

	waitErr := cmd.Wait()
	if interruptErr := <-errc; interruptErr != nil {
		return interruptErr
	}
	return waitErr
}

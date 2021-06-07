package helpers

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/osixia/container-baseimage/log"
)

// Exec wrapper
// =============================

type Exec struct {
	mu                  sync.Mutex
	process             *os.Process
	finished            bool
	cancellationSignals map[syscall.Signal]bool
	outputLogs          []*logWriter
	context             context.Context
	timeout             time.Duration
	setGPID             bool

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
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.Cmd == nil {
		return errors.New("Start called with a nil cmd")
	}

	if !e.noLog {
		log.Infof("Running %v ...", e.Cmd)
	}

	if err := e.prepare(); err != nil {
		return err
	}

	// CommandContext owns the initial cancellation; waitOrStop only escalates.
	stopSignal := os.Signal(os.Kill)
	if e.setGPID {
		stopSignal = syscall.SIGTERM
	} else if e.timeout > 0 {
		stopSignal = os.Interrupt
	}
	e.Cmd.Cancel = func() error {
		return e.cancelSignal(stopSignal)
	}
	e.Cmd.WaitDelay = e.timeout

	childProcesses.Lock()
	err := e.Cmd.Start()
	if err == nil {
		childProcesses.managed[e.Cmd.Process.Pid] = true
	}

	childProcesses.Unlock()
	if err != nil {
		return err
	}

	e.process = e.Cmd.Process
	e.finished = false
	pid := e.process.Pid

	if e.pidFile != "" {
		f, err := Create(e.pidFile)
		if err != nil {
			log.Warningf("Error creating pid file %v: %v", e.pidFile, err.Error())
		} else {
			defer func() {
				if err := f.Close(); err != nil {
					log.Warning(err.Error())
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

	if e.Cmd == nil {
		return errors.New("Wait called with a nil cmd ")
	}

	var err error
	if e.timeout > 0 {
		err = e.waitOrStop(e.context, e.Cmd, e.timeout)
	} else {
		err = e.Cmd.Wait()
	}

	for _, output := range e.outputLogs {
		output.Flush()
	}

	e.mu.Lock()
	e.finished = true
	childProcesses.Lock()
	if e.process != nil {
		delete(childProcesses.managed, e.process.Pid)
	}

	childProcesses.Unlock()
	e.mu.Unlock()

	if e.pidFile != "" {
		if err := Remove(e.pidFile); err != nil {
			log.Warningf("Error removing pid file %v: %v", e.pidFile, err.Error())
		}
	}

	err = e.cancellationError(err)

	if err != nil {
		err = fmt.Errorf("%v: %w", e.Cmd, err)
	}

	return err
}

// Signal sends sig to the process. If setGPID is true, the signal is sent to
// the entire process group so that child processes also receive it.
func (e *Exec) Signal(sig os.Signal) error {

	log.Tracef("Exec.Signal called with sig: %v", sig)

	e.mu.Lock()
	defer e.mu.Unlock()

	if e.process == nil || e.finished {
		return os.ErrProcessDone
	}

	return e.signal(sig)
}

func (e *Exec) signal(sig os.Signal) error {

	log.Tracef("Exec.signal called with sig: %v", sig)

	var err error

	if e.setGPID {

		s, ok := sig.(syscall.Signal)
		if !ok {
			return fmt.Errorf("unsupported signal type %T", sig)
		}

		err = syscall.Kill(-e.process.Pid, s)
	} else {
		err = e.process.Signal(sig)
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

	e.Cmd.Stdout = os.Stdout
	e.Cmd.Stderr = os.Stderr

	if e.stdoutLogPrintFunc != nil {
		output := &logWriter{print: e.stdoutLogPrintFunc}
		e.Cmd.Stdout = output
		e.outputLogs = append(e.outputLogs, output)
	}

	if e.stderrLogPrintFunc != nil {
		output := &logWriter{print: e.stderrLogPrintFunc}
		e.Cmd.Stderr = output
		e.outputLogs = append(e.outputLogs, output)
	}

	// set group id
	if e.setGPID {
		e.Cmd.SysProcAttr = &syscall.SysProcAttr{
			Setpgid: true,
		}
	}

	return nil
}

// cancelSignal records only signals successfully sent for context cancellation.
func (e *Exec) cancelSignal(sig os.Signal) error {

	log.Tracef("Exec.cancelSignal called with sig: %v", sig)

	e.mu.Lock()
	defer e.mu.Unlock()

	if e.process == nil || e.finished {
		return os.ErrProcessDone
	}

	err := e.signal(sig)
	if err == nil {
		if e.cancellationSignals == nil {
			e.cancellationSignals = make(map[syscall.Signal]bool)
		}
		e.cancellationSignals[sig.(syscall.Signal)] = true
	}

	return err
}

// Preserve normal exit failures even if the context was cancelled concurrently.
func (e *Exec) cancellationError(err error) error {

	log.Trace("Exec.cancellationError called")

	contextErr := e.context.Err()
	if contextErr == nil {
		return err
	}

	cancelled := errors.Is(err, contextErr)
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok && status.Signaled() {
			e.mu.Lock()
			cancelled = e.cancellationSignals[status.Signal()]
			// CommandContext may itself force-kill the child after WaitDelay.
			if !cancelled && status.Signal() == syscall.SIGKILL && e.timeout > 0 {
				cancelled = len(e.cancellationSignals) > 0
			}
			e.mu.Unlock()
		}
	}

	if cancelled {
		if contextErr == context.Canceled {
			return nil
		}
		return contextErr
	}

	return err
}

// waitOrStop lets CommandContext send the initial signal, then escalates after
// killDelay. Always return Wait's result so a concurrent exit failure survives.
func (e *Exec) waitOrStop(ctx context.Context, cmd *exec.Cmd, killDelay time.Duration) (waitErr error) {

	log.Tracef("Exec.waitOrStop called with cmd: %v, killDelay: %v", cmd, killDelay)

	if cmd.Process == nil {
		return errors.New("waitOrStop called with a nil cmd.Process — missing Start call?")
	}

	// WaitDelay can kill only the group leader. Finish killing its group
	// if it fires before our escalation timer, including inherited output pipes.
	defer func() {
		if ctx.Err() == nil || !e.setGPID {
			return
		}

		forced := errors.Is(waitErr, exec.ErrWaitDelay)
		var exitErr *exec.ExitError
		if errors.As(waitErr, &exitErr) {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				forced = status.Signaled() && status.Signal() == syscall.SIGKILL
			}
		}

		if forced {
			if err := e.cancelSignal(os.Kill); err != nil && !errors.Is(err, os.ErrProcessDone) {
				log.Warningf("Failed to force-kill command group %v: %v", cmd, err)
			}
		}
	}()

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
	}

	timer := time.NewTimer(killDelay)
	defer timer.Stop()

	select {
	case err := <-done:
		return err
	case <-timer.C:
		if err := e.cancelSignal(os.Kill); err != nil && !errors.Is(err, os.ErrProcessDone) {
			log.Warningf("Failed to force-kill command %v: %v", cmd, err)
		}
	}

	return <-done
}

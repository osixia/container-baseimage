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
	cancellationTime    time.Time
	logWriters          []*log.Writer
	context             context.Context
	timeout             time.Duration
	setGPID             bool
	stopGroupOnExit     bool
	stopSignal          os.Signal
	startedStopSignal   os.Signal

	Cmd     *exec.Cmd
	pidFile string

	env []string

	stdoutLogPrintFunc log.PrintFunc
	stderrLogPrintFunc log.PrintFunc

	noLog   bool
	logArgs bool
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

// WithStopGroupOnExit stops remaining children even when the leader exits first.
func (e *Exec) WithStopGroupOnExit(b bool) *Exec {

	e.stopGroupOnExit = b

	return e
}

// WithStopSignal overrides the initial signal used for context cancellation.
// A zero shutdown delay still forces an immediate kill.
func (e *Exec) WithStopSignal(sig os.Signal) *Exec {

	e.stopSignal = sig
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

// WithLogArgs also logs the command arguments at info level and in errors. Only
// for commands built by the container binary whose arguments are never secrets
// (e.g. package names): user provided commands may hold passwords.
func (e *Exec) WithLogArgs(b bool) *Exec {

	e.logArgs = b

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
		log.Infof("Running %v ...", e.logName())
	}

	if err := e.prepare(); err != nil {
		return err
	}

	// CommandContext owns the initial cancellation; waitOrStop only escalates.
	stopSignal := os.Signal(os.Kill)
	if e.setGPID && e.timeout > 0 {
		stopSignal = syscall.SIGTERM
	} else if e.timeout > 0 {
		stopSignal = os.Interrupt
	}
	if e.timeout > 0 && e.stopSignal != nil {
		stopSignal = e.stopSignal
	}
	e.startedStopSignal = stopSignal
	e.Cmd.Cancel = func() error {
		return e.cancelSignal(stopSignal)
	}
	e.Cmd.WaitDelay = max(e.timeout, 0)

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
		if err := WriteFileAtomic(e.pidFile, []byte(strconv.Itoa(pid)), 0644, 0755); err != nil {
			log.Warningf("Error writing pid file %v: %v", e.pidFile, err)
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
	if e.timeout > 0 || e.setGPID || e.stopSignal != nil {
		err = e.waitOrStop(e.context, e.Cmd, max(e.timeout, 0))
	} else {
		err = e.Cmd.Wait()
	}

	for _, writer := range e.logWriters {
		writer.Flush()
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
		err = fmt.Errorf("%v: %w", e.logName(), err)
	}

	return err
}

// Signal forwards sig to the process only, even with setGPID: like an init,
// the process handles its children (e.g. a daemon reloading on SIGHUP). Stop
// signals sent on cancellation still reach the entire process group.
func (e *Exec) Signal(sig os.Signal) error {

	log.Tracef("Exec.Signal called with sig: %v", sig)

	e.mu.Lock()
	defer e.mu.Unlock()

	if e.process == nil || e.finished {
		return os.ErrProcessDone
	}

	err := e.process.Signal(sig)
	if errors.Is(err, os.ErrProcessDone) || errors.Is(err, syscall.ESRCH) {
		return os.ErrProcessDone
	}

	return err
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

// logName is the command shown at info level and in errors: arguments are
// only shown at debug level unless WithLogArgs is set.
func (e *Exec) logName() string {

	if e.logArgs {
		return e.Cmd.String()
	}

	return e.Cmd.Path
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
		output := log.NewWriter(e.stdoutLogPrintFunc)
		e.Cmd.Stdout = output
		e.logWriters = append(e.logWriters, output)
	}

	if e.stderrLogPrintFunc != nil {
		output := log.NewWriter(e.stderrLogPrintFunc)
		e.Cmd.Stderr = output
		e.logWriters = append(e.logWriters, output)
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
		if e.cancellationTime.IsZero() {
			e.cancellationTime = time.Now()
		}
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
			// A terminal may deliver the stop signal itself, before cancellation sends it.
			cancelled = e.cancellationSignals[status.Signal()] || status.Signal() == e.startedStopSignal
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
func (e *Exec) waitOrStop(ctx context.Context, cmd *exec.Cmd, killDelay time.Duration) error {

	log.Tracef("Exec.waitOrStop called with cmd: %v, killDelay: %v", cmd, killDelay)

	if cmd.Process == nil {
		return errors.New("waitOrStop called with a nil cmd.Process — missing Start call?")
	}

	// Even if Wait returns early, descendants retain the original stop deadline.
	defer e.stopRemainingGroup(killDelay)

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
			log.Warningf("Failed to force-kill command %v: %v", e.logName(), err)
		}
	}

	return <-done
}

// stopRemainingGroup keeps the grace period after the leader exits. Services
// also stop remaining children after a natural exit, before they can restart.
func (e *Exec) stopRemainingGroup(killDelay time.Duration) {

	e.mu.Lock()
	stopped := !e.cancellationTime.IsZero()
	if !e.setGPID || (!stopped && !e.stopGroupOnExit) {
		e.mu.Unlock()
		return
	}
	pid := e.process.Pid
	stopSignal := e.startedStopSignal.(syscall.Signal)
	deadline := e.cancellationTime.Add(killDelay)
	forced := e.cancellationSignals[syscall.SIGKILL]
	e.mu.Unlock()

	if !stopped {
		if err := syscall.Kill(-pid, stopSignal); err != nil {
			if !errors.Is(err, syscall.ESRCH) {
				log.Warningf("Failed to stop command group %v: %v", pid, err)
			}
			return
		}
		deadline = time.Now().Add(killDelay)
		forced = stopSignal == syscall.SIGKILL
	}

	timer := time.NewTimer(max(time.Until(deadline), 0))
	defer timer.Stop()
	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()

	for {
		err := syscall.Kill(-pid, 0)
		if errors.Is(err, syscall.ESRCH) {
			return
		}
		if err != nil && !errors.Is(err, syscall.EPERM) {
			log.Warningf("Failed to check command group %v: %v", pid, err)
			return
		}

		if forced || !time.Now().Before(deadline) {
			if err := syscall.Kill(-pid, syscall.SIGKILL); err != nil && !errors.Is(err, syscall.ESRCH) {
				log.Warningf("Failed to force-kill command group %v: %v", pid, err)
			}
			return
		}

		select {
		case <-ticker.C:
		case <-timer.C:
		}
	}
}

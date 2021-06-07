package core

import (
	"context"
	goerrors "errors"
	"fmt"
	"os"
	"os/signal"
	"slices"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/google/shlex"
	"golang.org/x/sync/errgroup"

	"github.com/osixia/container-baseimage/errors"
	"github.com/osixia/container-baseimage/helpers"
	"github.com/osixia/container-baseimage/log"
)

// Lifecycle steps constantes and global variables
// =============================

const (
	LifecycleStepStartup LifecycleStep = "startup"
	LifecycleStepProcess LifecycleStep = "process"
	LifecycleStepFinish  LifecycleStep = "finish"
)

var LifecycleSteps = []LifecycleStep{
	LifecycleStepStartup,
	LifecycleStepProcess,
	LifecycleStepFinish,
}

var LifecycleInterceptedSignals = []os.Signal{
	syscall.SIGINT,
	syscall.SIGTERM,
	syscall.SIGQUIT,
}

// Lifecycle steps
// =============================

type LifecycleStep string

func LifecycleStepsList() []string {

	values := make([]string, 0, len(LifecycleSteps))
	for _, f := range LifecycleSteps {
		values = append(values, string(f))
	}

	return values
}

// Lifecycle options
// =============================

type LifecycleOptions struct {
	SkipStartup bool
	SkipProcess bool
	SkipFinish  bool

	PreStartupCmds []string
	PreProcessCmds []string
	PreFinishCmds  []string
	PreExitCmds    []string

	Args    []string
	RunBash bool

	// run bash instead of services process scripts, keeping startup and finish
	SkipProcessServices bool

	KillAllOnExit        bool
	KillAllOnExitTimeout time.Duration
	RestartProcesses     *bool
	NoRestartDelay       bool
}

// Lifecycle
// =============================

type lifecycle struct {
	prcs Processes

	startupServices []Service
	processServices []Service
	finishServices  []Service

	step     atomic.Value
	exitCode int

	processes []*lifecycleProcess

	// stdin is a terminal (docker run -it)
	terminal bool

	options *LifecycleOptions
}

func newLifecycle(prcs Processes, lco *LifecycleOptions, ss []Service) *lifecycle {

	var sss, pss, fss []Service

	for _, s := range ss {
		if s.StartupFile() != "" {
			sss = append(sss, s)
		}
		if s.ProcessFile() != "" && !lco.SkipProcessServices {
			pss = append(pss, s)
		}
		if s.FinishFile() != "" {
			fss = append(fss, s)
		}
	}

	return &lifecycle{
		prcs: prcs,

		startupServices: sss,
		processServices: pss,
		finishServices:  fss,

		terminal: helpers.IsTerminal(os.Stdin),

		options: lco,
	}
}

func (lc *lifecycle) ExitCode() int {
	return lc.exitCode
}

func (lc *lifecycle) exec(ctx context.Context) *helpers.Exec {
	return helpers.NewExec(ctx).WithTimeout(lc.options.KillAllOnExitTimeout)
}

// commandExec runs startup, finish, pre and custom commands. Without a terminal,
// the command and its children are stopped with SIGTERM. On a terminal, it must
// stay in the foreground group, where Ctrl+C already sends SIGINT: stop with it too.
func (lc *lifecycle) commandExec(ctx context.Context) *helpers.Exec {

	if lc.terminal {
		return lc.exec(ctx).WithStopSignal(os.Interrupt)
	}
	return lc.exec(ctx).WithStopSignal(syscall.SIGTERM).WithSetGPID(true)
}

func (lc *lifecycle) run(ctx context.Context, keepAlive bool) {

	log.Trace("lifecycle.Run called")

	runCtx, cancelRun := context.WithCancel(ctx)

	processCtx, cancelProcess := context.WithCancel(runCtx)
	defer cancelProcess()

	// Go exits on an unhandled SIGHUP: keep it handled for the whole run.
	// During the process step it is still forwarded to the services.
	hangup := make(chan os.Signal, 1)
	signal.Notify(hangup, syscall.SIGHUP)
	defer signal.Stop(hangup)

	// catch first interrupt
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, LifecycleInterceptedSignals...)

	go func() {
		for {
			select {
			case <-interrupt:
				log.Info("Container execution aborted (SIGINT, SIGTERM or SIGQUIT signal received)")
				if lc.step.Load() == LifecycleStepProcess {
					log.Trace("cancel process context")
					cancelProcess()
				} else {
					log.Trace("cancel run context")
					cancelRun()
				}
			case <-runCtx.Done():
				signal.Stop(interrupt)
				return
			}
		}
	}()

	// Only PID 1 adopts orphan processes. Reap unmanaged children without
	// consuming the exit statuses of commands supervised by helpers.Exec.
	reaperCtx, stopReaper := context.WithCancel(context.WithoutCancel(ctx))
	reaperDone := make(chan struct{})

	if os.Getpid() == 1 {
		go func() {
			defer close(reaperDone)

			helpers.ReapChildren(reaperCtx)
		}()
	} else {
		close(reaperDone)
	}

	defer func() {
		keep := keepAlive && processCtx.Err() == nil

		// Registered before exit cancels the interrupt handler, so no shutdown
		// signal is left unhandled. A signal received during exit skips keep alive.
		shutdown := make(chan os.Signal, 1)
		signal.Notify(shutdown, LifecycleInterceptedSignals...)
		defer signal.Stop(shutdown)

		lc.exit(ctx, cancelRun)

		// Keep signal handling and orphan reaping active until an actual shutdown.
		if keep {
			log.Info("All processes have exited, keep container alive ☠ ...")
			select {
			case <-shutdown:
			case <-ctx.Done():
			}
		}

		stopReaper()
		<-reaperDone
	}()

	// startup
	if !lc.options.SkipStartup {
		if err := lc.runStep(runCtx, LifecycleStepStartup, lc.options.PreStartupCmds, lc.startup); err != nil {
			lc.exitCode = errors.ExitCode(err)

			log.Error(err.Error())
			return
		}
	}

	// process
	if !lc.options.SkipProcess {
		if err := lc.runStep(processCtx, LifecycleStepProcess, lc.options.PreProcessCmds, lc.process); err != nil {
			lc.exitCode = errors.ExitCode(err)

			log.Error(err.Error())
		}
	}

	// finish
	if !lc.options.SkipFinish {
		if err := lc.runStep(runCtx, LifecycleStepFinish, lc.options.PreFinishCmds, lc.finish); err != nil {
			if lc.exitCode == 0 {
				lc.exitCode = errors.ExitCode(err)
			}

			log.Error(err.Error())
			return
		}
	}
}

func (lc *lifecycle) runStep(ctx context.Context, step LifecycleStep, preCmds []string, stepFunc func(context.Context) error) error {

	log.Trace("lifecycle.runStep called")

	if ctx.Err() == context.Canceled {
		log.Debugf("Ignoring %v lifecycle step (container execution aborted) ...", step)
		return nil
	}

	lc.step.Store(step)

	log.Debugf("Starting %v lifecycle step ...", step)
	if err := lc.execPreCommands(ctx, preCmds, string(step)); err != nil {
		return err
	}

	return stepFunc(ctx)
}

func (lc *lifecycle) startup(ctx context.Context) error {

	log.Trace("lifecycle.startup called")

	for _, s := range lc.startupServices {
		// Stop requested: do not start the next command.
		if ctx.Err() != nil {
			return nil
		}

		if err := lc.commandExec(ctx).Command(s.StartupFile()).Run(); err != nil {
			return err
		}
	}

	return nil
}

func (lc *lifecycle) process(ctx context.Context) error {

	log.Trace("lifecycle.process called")

	processCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	g, subCtx := errgroup.WithContext(processCtx)

	// Install the watch before starting services so a stop request cannot
	// be lost between launching a child and registering its directory watch.
	var watcher *fsnotify.Watcher
	if len(lc.processServices) > 0 {
		var err error
		watcher, err = lc.prcs.NewWatcher()
		if err != nil {
			return err
		}

		defer func() {
			if err := watcher.Close(); err != nil {
				log.Warning(err.Error())
			}
		}()
	}

	// Register before launching commands so their first signal can be queued.
	sigChan := make(chan os.Signal, 32)
	signal.Notify(sigChan)
	defer func() {
		signal.Stop(sigChan)
		close(sigChan)
	}()

	lc.runProcessCommand(subCtx, g, cancel)
	lc.runProcessServices(subCtx, g)

	if len(lc.processes) > 0 {
		go lc.forwardSignals(sigChan)
		if watcher != nil {
			go lc.watchProcesses(watcher)
		}
	}

	return g.Wait()
}

type lifecycleProcess struct {
	name    string
	process Process
	mu      sync.RWMutex
	exec    *helpers.Exec
	cancel  context.CancelFunc
}

func (lc *lifecycle) runProcessServices(ctx context.Context, g *errgroup.Group) {

	log.Trace("lifecycle.runProcessServices called")

	// do not restart services by default
	restart := false

	// if option is set, set restart accordingly
	if lc.options.RestartProcesses != nil {
		restart = *lc.options.RestartProcesses
	}

	// if option is not set, and this is a multi-process set restart to true
	if lc.options.RestartProcesses == nil && len(lc.processServices) > 1 {
		restart = true
	}

	for _, s := range lc.processServices {

		s := s

		log.Tracef("prepare running service %v", s.Name())

		p, err := lc.prcs.New(s)
		if err != nil {
			g.Go(func() error {
				return fmt.Errorf("%v: preparing process: %w", s.Name(), err)
			})
			continue
		}

		lcp := &lifecycleProcess{
			name:    s.Name(),
			process: p,
		}

		lc.processes = append(lc.processes, lcp)

		stopSignal := s.StopSignal()

		setStarting := func(b bool) {
			if err := p.SetStarting(b); err != nil {
				log.Warningf("%v: %v", s.Name(), err)
			}
		}

		g.Go(func() error {
			defer setStarting(false)

			delay := 100 * time.Millisecond
			for {

				if ctx.Err() != nil {
					return nil
				}

				if p.IsWantedDown() {
					log.Tracef("%v: wanted down", s.Name())
					setStarting(false)
					if !waitForRestart(ctx, time.Second) {
						return nil
					}

					continue
				}

				subCtx, cancelCtx := context.WithCancel(ctx)

				lcp.mu.Lock()
				lcp.cancel = cancelCtx
				exec := lc.exec(subCtx).WithSetGPID(true).WithStopGroupOnExit(true).WithPIDFile(p.PIDFile()).Command(s.ProcessFile(), lc.options.Args...)
				if stopSignal != 0 {
					exec.WithStopSignal(stopSignal)
				}
				lcp.exec = exec
				lcp.mu.Unlock()

				// A stop request may have been handled before the cancellation callback was set.
				if p.IsWantedDown() {
					cancelCtx()
					continue
				}

				started := time.Now()
				err := exec.Start()
				// the pid file is written: the process is no longer starting
				setStarting(false)
				if err == nil {
					err = exec.Wait()
				}

				// Read before cleanup: start may already have removed the stop marker.
				stopped := subCtx.Err() != nil
				cancelCtx()

				if err != nil {
					err = fmt.Errorf("%v: %w", s.Name(), err)
				}

				if ctx.Err() != nil {
					return err
				}

				if stopped || p.IsWantedDown() {
					log.Infof("%v: stopped on request, checking desired state ...", s.Name())
					continue
				}

				if !restart {
					log.Tracef("%v: ended restart disabled", s.Name())

					return err
				}

				// log why the process ended, the error is not returned when restarting
				if err != nil {
					log.Warningf("%v: restarting ...", err)
				} else {
					log.Infof("%v: ended, restarting ...", s.Name())
				}

				// keep the process listed until it is restarted
				setStarting(true)

				if lc.options.NoRestartDelay {
					continue
				}

				if time.Since(started) >= 30*time.Second {
					delay = 100 * time.Millisecond
				}

				if !waitForRestart(ctx, delay) {
					return nil
				}

				delay = min(delay*2, 5*time.Second)
			}
		})
	}
}

func (lc *lifecycle) watchProcesses(w *fsnotify.Watcher) {

	log.Trace("lifecycle.watchProcesses called")

	for {
		select {

		case err, ok := <-w.Errors:

			if !ok { // channel was closed
				return
			}

			log.Warningf("Process watcher failed: %v", err)
			// Reconcile after an error: an overflow can have dropped a down event.
			for _, lcp := range lc.processes {
				lcp.stopIfWantedDown()
			}

		case e, ok := <-w.Events:

			if !ok { // channel was closed
				return
			}

			log.Tracef("recieved watch event %+v", e)

			for _, lcp := range lc.processes {
				if lcp.process == nil || e.Name != lcp.process.WantedDownFile() {
					continue
				}
				// Read the current state; queued events can predate a start request.
				lcp.stopIfWantedDown()
				break
			}
		}
	}
}

// The callback and desired state must refer to the same process generation.
func (lcp *lifecycleProcess) stopIfWantedDown() {

	if lcp.process == nil {
		return
	}
	lcp.mu.RLock()
	defer lcp.mu.RUnlock()
	if lcp.cancel != nil && lcp.process.IsWantedDown() {
		log.Tracef("cancelling %v process context", lcp.name)
		lcp.cancel()
	}
}

func (lc *lifecycle) forwardSignals(sigChan <-chan os.Signal) {

	log.Trace("lifecycle.forwardSignals called")

	processes := make([]Signaler, 0, len(lc.processes))
	for _, lcp := range lc.processes {
		processes = append(processes, lcp)
	}

	ForwardSignals(sigChan, processes...)
}

// Signal sends sig to the current exec of the process, if any.
func (lcp *lifecycleProcess) Signal(sig os.Signal) error {

	lcp.mu.RLock()
	exec := lcp.exec
	lcp.mu.RUnlock()

	if exec == nil {
		return os.ErrProcessDone
	}

	if err := exec.Signal(sig); err != nil {
		return fmt.Errorf("process %v: %w", lcp.name, err)
	}

	return nil
}

// Signaler is a process signals can be forwarded to.
type Signaler interface {
	Signal(sig os.Signal) error
}

// ForwardSignals forwards the signals received on sigChan to processes until
// sigChan is closed. Intercepted signals and SIGCHLD are not forwarded.
func ForwardSignals(sigChan <-chan os.Signal, processes ...Signaler) {

	log.Trace("ForwardSignals called")

	ignoreSignals := append(LifecycleInterceptedSignals, syscall.SIGCHLD)

	for sig := range sigChan {

		// The Go runtime sends itself SIGURG for goroutine preemption, up to
		// hundreds per second: never forward nor log them.
		if sig == syscall.SIGURG {
			continue
		}

		ignore := slices.Contains(ignoreSignals, sig)

		if ignore {
			log.Debugf("Ignoring signal: %v", sig)
			continue
		}

		log.Debugf("Sending %v signal to child processes ...", sig)

		for _, p := range processes {
			if err := p.Signal(sig); err != nil && !goerrors.Is(err, os.ErrProcessDone) {
				log.Warningf("Error forwarding %v: %v", sig, err)
			}
		}
	}
}

// runProcessCommand runs the custom command or bash: when it exits (e.g. Ctrl+D),
// whatever its exit code, the process step ends and services are stopped.
func (lc *lifecycle) runProcessCommand(ctx context.Context, g *errgroup.Group, stop context.CancelFunc) {

	log.Trace("lifecycle.runProcessCommand called")

	var cmds [][]string

	runBash := lc.options.RunBash

	// no service to run
	if len(lc.processServices) == 0 {

		// empty command line: force to run bash
		if len(lc.options.Args) == 0 {
			runBash = true
		} else {
			// else run command line
			cmds = append(cmds, lc.options.Args)
		}
	}

	// add bash to commands to run
	if runBash {
		cmds = append(cmds, []string{"bash"})
	}

	for _, cmd := range cmds {
		exec := lc.commandExec(ctx).Command(cmd[0], cmd[1:]...)
		lcp := &lifecycleProcess{name: cmd[0], exec: exec}
		lc.processes = append(lc.processes, lcp)

		g.Go(func() error {
			defer stop()
			return exec.Run()
		})
	}
}

func (lc *lifecycle) finish(ctx context.Context) error {

	log.Trace("lifecycle.finish called")

	for _, s := range lc.finishServices {
		// Stop requested: do not start the next command.
		if ctx.Err() != nil {
			return nil
		}

		if err := lc.commandExec(ctx).Command(s.FinishFile()).Run(); err != nil {
			return err
		}
	}

	return nil
}

func (lc *lifecycle) exit(ctx context.Context, ctxCancelFunc context.CancelFunc) {

	log.Tracef("lifecycle.exit called with ctxCancelFunc: %v", ctxCancelFunc)

	// Keep exit commands interruptible after cancelling the lifecycle context.
	ctx, stop := signal.NotifyContext(ctx, LifecycleInterceptedSignals...)
	defer stop()

	log.Trace("Calling ctxCancelFunc ...")
	ctxCancelFunc()

	log.Debug("Starting exit lifecycle step ...")
	if err := lc.execPreCommands(ctx, lc.options.PreExitCmds, "exit"); err != nil {
		if lc.exitCode == 0 {
			lc.exitCode = errors.ExitCode(err)
		}

		log.Error(err.Error())
	}

	if lc.options.KillAllOnExit {
		lc.killAll()
	}

	log.Info("Exiting ...")
}

func (lc *lifecycle) killAll() {

	log.Trace("lifecycle.killAll called")

	// security to not kill all processes if executable is run outside a container
	if os.Getpid() != 1 {
		log.Warning("Current process is not pid 1: ignoring terminating all processes ...")
		return
	}

	// if no others proccess is running return
	if pids, err := helpers.ListPIDs(); err != nil {
		log.Errorf("Error listing remaining processes: %v", err)
		return
	} else if len(pids) == 0 {
		return
	}

	timeout := max(lc.options.KillAllOnExitTimeout, 0)
	log.Infof("Terminating all processes (timeout: %v) ...", timeout)

	if err := helpers.KillAll(syscall.SIGTERM); err != nil {
		log.Errorf("Error terminating all processes: %v", err)
	}

	// Bound both the graceful wait and the wait for reaping after SIGKILL.
	killDeadline := time.Now().Add(timeout)
	exitDeadline := killDeadline.Add(time.Second)
	killed := false

	for {
		pids, err := helpers.ListPIDs()
		if err != nil {
			log.Errorf("Error listing remaining processes: %v", err)
			return
		}

		if len(pids) == 0 {
			return
		}

		now := time.Now()
		if !killed && !now.Before(killDeadline) {
			log.Info("Terminating all processes: timeout reached, killing all processes ...")
			if err := helpers.KillAll(syscall.SIGKILL); err != nil {
				log.Errorf("Error killing all processes: %v", err)
			}
			killed = true
		}

		if !now.Before(exitDeadline) {
			log.Warningf("Stopping cleanup with remaining processes: %v", pids)
			return
		}

		log.Debugf("%v child processes still running ...", len(pids))
		log.Tracef("child processes: %v ...", pids)

		nextDeadline := exitDeadline
		if !killed {
			nextDeadline = killDeadline
		}
		time.Sleep(min(250*time.Millisecond, time.Until(nextDeadline)))
	}

}

func (lc *lifecycle) execPreCommands(ctx context.Context, commands []string, step string) error {

	log.Tracef("lifecycle.execPreCommands called with commands: %v, step: %v", commands, step)

	if len(commands) == 0 {
		return nil
	}

	log.Infof("Running pre-%v commands...", step)

	for _, command := range commands {

		// Stop requested: do not start the next command.
		if ctx.Err() != nil {
			return nil
		}

		shCmd, err := shlex.Split(command)
		if err != nil {
			return err
		}

		if len(shCmd) == 0 {
			continue
		}

		if err := lc.commandExec(ctx).Command(shCmd[0], shCmd[1:]...).Run(); err != nil {
			return err
		}
	}

	return nil
}

func waitForRestart(ctx context.Context, delay time.Duration) bool {

	log.Tracef("waitForRestart called with delay: %v", delay)

	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():

		return false
	case <-timer.C:

		return true
	}
}

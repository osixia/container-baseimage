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
	"github.com/hashicorp/go-reap"
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

	TerminateAllOnExit        bool
	TerminateAllOnExitTimeout time.Duration
	RestartProcesses          *bool
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

	options *LifecycleOptions
}

func newLifecycle(prcs Processes, lco *LifecycleOptions, ss []Service) *lifecycle {

	var sss, pss, fss []Service

	for _, s := range ss {
		if s.StartupFile() != "" {
			sss = append(sss, s)
		}
		if s.ProcessFile() != "" {
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

		options: lco,
	}
}

func (lc *lifecycle) ExitCode() int {
	return lc.exitCode
}

func (lc *lifecycle) exec(ctx context.Context) *helpers.Exec {
	return helpers.NewExec(ctx).WithTimeout(lc.options.TerminateAllOnExitTimeout)
}

func (lc *lifecycle) run(ctx context.Context) {

	log.Trace("lifecycle.Run called")

	runCtx, cancelRun := context.WithCancel(ctx)
	defer lc.exit(ctx, cancelRun) // exit

	processCtx, cancelProcess := context.WithCancel(runCtx)

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

	// reap zombie processes
	pids := make(reap.PidCh, 1)
	errs := make(reap.ErrorCh, 1)
	done := make(chan struct{})

	var reapLock sync.RWMutex
	go reap.ReapChildren(pids, errs, done, &reapLock)

	go func() {
		log.Trace("Starting zombie process reaper ...")
		for {
			select {
			case pid, ok := <-pids:
				if !ok {
					continue
				}
				log.Debugf("Reaped pid: %v", pid)

			case e, ok := <-errs:
				if !ok {
					continue
				}
				log.Debugf("Failed to reap zombie process: %v", e)
			}
		}
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
		if err := lc.exec(ctx).Command(s.StartupFile()).Run(); err != nil {
			return err
		}
	}

	return nil
}

func (lc *lifecycle) process(ctx context.Context) error {

	log.Trace("lifecycle.process called")

	g, subCtx := errgroup.WithContext(ctx)

	// run commands
	lc.runProcessCommand(subCtx, g)

	// run services
	lc.runProcessServices(subCtx, g)

	if len(lc.processes) > 0 {

		// forward signals to processes
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan)

		go lc.forwardSignals(sigChan)

		// watch processes files
		w, err := lc.prcs.NewWatcher()
		if err != nil {
			return err
		}
		defer func() {
			if err := w.Close(); err != nil {
				log.Error(err.Error())
			}
		}()

		go lc.watchProcesses(w)
	}

	return g.Wait()
}

type lifecycleProcess struct {
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
			log.Warning(err.Error())
			continue
		}

		lcp := &lifecycleProcess{
			process: p,
		}

		lc.processes = append(lc.processes, lcp)

		g.Go(func() error {

			for {

				if ctx.Err() == context.Canceled {
					log.Tracef("%v: context cancelled", s.Name())
					break
				}

				if p.IsWantedDown() {
					log.Tracef("%v: wanted down", s.Name())
					time.Sleep(1 * time.Second)
					continue
				}

				subCtx, cancelCtx := context.WithCancel(ctx)

				lcp.mu.Lock()
				lcp.cancel = cancelCtx
				exec := lc.exec(subCtx).WithSetGPID(true).WithPIDFile(p.PIDFile()).Command(s.ProcessFile(), lc.options.Args...)
				lcp.exec = exec
				lcp.mu.Unlock()

				err = exec.Run()
				cancelCtx()

				if err != nil {
					err = fmt.Errorf("%v: %w", s.Name(), err)
				}

				if p.IsWantedDown() {
					log.Infof("%v: intentionally down — awaiting desired up state", s.Name())
					continue
				}

				if !restart {
					log.Tracef("%v: ended restart disabled", s.Name())
					break
				}
			}

			return err
		})
	}
}

func (lc *lifecycle) watchProcesses(w *fsnotify.Watcher) {

	log.Trace("lifecycle.watchProcesses called")

	for {
		select {

		case _, ok := <-w.Errors:

			if !ok { // channel was closed
				return
			}

		case e, ok := <-w.Events:

			if !ok { // channel was closed
				return
			}

			log.Tracef("recieved watch event %+v", e)

			if e.Op&fsnotify.Create == fsnotify.Create {

				for _, lcp := range lc.processes {
					if e.Name != lcp.process.WantedDownFile() {
						continue
					}

					log.Tracef("%v process %v detected", lcp.process.Name(), lcp.process.WantedDownFile())

					lcp.mu.RLock()
					exec, cancel := lcp.exec, lcp.cancel
					lcp.mu.RUnlock()

					if exec != nil && cancel != nil {
						log.Tracef("cancelling %v process context", lcp.process.Name())
						cancel()
					}

					break
				}
			}
		}
	}
}

func (lc *lifecycle) forwardSignals(sigChan <-chan os.Signal) {

	log.Trace("lifecycle.forwardSignals called")

	ignoreSignals := append(LifecycleInterceptedSignals, syscall.SIGCHLD)

	for sig := range sigChan {

		ignore := slices.Contains(ignoreSignals, sig)

		if ignore {
			log.Debugf("Ignoring signal: %v", sig)
			continue
		}

		log.Debugf("Sending %v signal to childs processes ...", sig)

		for _, lcp := range lc.processes {
			lcp.mu.RLock()
			exec := lcp.exec
			lcp.mu.RUnlock()

			if exec == nil || exec.Cmd == nil || exec.Cmd.Process == nil {
				continue
			}

			if exec.Cmd.ProcessState != nil && exec.Cmd.ProcessState.Exited() {
				continue
			}

			if err := exec.Signal(sig); err != nil && !goerrors.Is(err, os.ErrProcessDone) {
				log.Warningf("Error during %v signal transmission to pid %v: %v", sig, exec.Cmd.Process.Pid, err)
			}
		}
	}
}

func (lc *lifecycle) runProcessCommand(ctx context.Context, g *errgroup.Group) {

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
		cmd := cmd

		g.Go(func() error {
			exec := lc.exec(ctx).Command(cmd[0], cmd[1:]...)
			return exec.Run()
		})
	}
}

func (lc *lifecycle) finish(ctx context.Context) error {

	log.Trace("lifecycle.finish called")

	for _, s := range lc.finishServices {
		if err := lc.exec(ctx).Command(s.FinishFile()).Run(); err != nil {
			return err
		}
	}

	return nil
}

func (lc *lifecycle) exit(ctx context.Context, ctxCancelFunc context.CancelFunc) {

	log.Tracef("lifecycle.exit called with ctxCancelFunc: %v", ctxCancelFunc)

	log.Trace("Calling ctxCancelFunc ...")
	ctxCancelFunc()

	log.Debug("Starting exit lifecycle step ...")
	if err := lc.execPreCommands(ctx, lc.options.PreExitCmds, "exit"); err != nil {
		lc.exitCode = errors.ExitCode(err)

		log.Error(err.Error())
	}

	if lc.options.TerminateAllOnExit {
		lc.killAll()
	}

	log.Info("Exiting ...")
}

func (lc *lifecycle) killAll() {

	log.Trace("lifecycle.killAll called")

	// security to not kill all processes if executable is run outside a container
	if os.Getpid() != 1 {
		log.Debug("Current process is not pid 1: ignoring terminating all processes ...")
		return
	}

	// if no others proccess is running return
	if pids, _ := helpers.ListPIDs(); len(pids) == 0 {
		return
	}

	timeout := lc.options.TerminateAllOnExitTimeout
	log.Infof("Terminating all processes (timeout: %v) ...", timeout)

	if err := helpers.KillAll(syscall.SIGTERM); err != nil {
		log.Errorf("Error terminating all processes: %v", err.Error())
	}

	timer := time.AfterFunc(timeout, func() {
		log.Info("Terminating all processes: timeout reached, killing all processes ...")
		if err := helpers.KillAll(syscall.SIGKILL); err != nil {
			log.Errorf("Error killing all processes: %v", err.Error())
		}
	})
	defer timer.Stop()

	for {
		pids, _ := helpers.ListPIDs()
		childs := len(pids)
		if childs == 0 {
			break
		}
		log.Debugf("%v child processes still running ...", childs)
		log.Tracef("child processes: %v ...", pids)
		time.Sleep(250 * time.Millisecond)
	}
}

func (lc *lifecycle) execPreCommands(ctx context.Context, commands []string, step string) error {

	log.Tracef("lifecycle.execPreCommands called with commands: %v, step: %v", commands, step)

	if len(commands) == 0 {
		return nil
	}

	log.Infof("Running pre-%v commands...", step)

	for _, command := range commands {

		shCmd, err := shlex.Split(command)
		if err != nil {
			return err
		}

		if len(shCmd) == 0 {
			continue
		}

		if err := lc.exec(ctx).Command(shCmd[0], shCmd[1:]...).Run(); err != nil {
			return err
		}
	}

	return nil
}

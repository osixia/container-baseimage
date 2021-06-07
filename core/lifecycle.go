package core

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/osixia/container-baseimage/log"
	"golang.org/x/sync/errgroup"
)

const (
	LifecycleStepStartup LifecycleStep = "startup"
	LifecycleStepProcess LifecycleStep = "process"
	LifecycleStepFinish  LifecycleStep = "finish"
	LifecycleStepExit    LifecycleStep = "exit"
)

var LifecycleSteps = []LifecycleStep{
	LifecycleStepStartup,
	LifecycleStepProcess,
	LifecycleStepFinish,
	LifecycleStepExit,
}

type LifecycleStep string

func (epls LifecycleStep) String() string {
	return string(epls)
}

type LifecycleOptions struct {
	SkipStartup bool
	SkipProcess bool
	SkipFinish  bool

	PreStartupCmds []string
	PreProcessCmds []string
	PreFinishCmds  []string
	PreExitCmds    []string

	Commands []string
	RunBash  bool

	TerminateAllOnExit        bool
	TerminateAllOnExitTimeout time.Duration
	RestartProcesses          bool
}

type lifecycleConfig struct {
	startupScripts []string

	processDir     string
	processScripts []string

	finishScripts []string
}

func newLifecycleConfig(lss []EntrypointLinkedService, pd string) *lifecycleConfig {

	sss := []string{}
	pss := []string{}
	fss := []string{}

	for _, ls := range lss {
		ss := ls.Script(LifecycleStepStartup)
		ps := ls.Script(LifecycleStepProcess)
		fs := ls.Script(LifecycleStepFinish)

		if ss != "" {
			sss = append(sss, ss)
		}

		if ps != "" {
			pss = append(pss, ps)
		}

		if fs != "" {
			fss = append(fss, fs)
		}

	}

	return &lifecycleConfig{
		startupScripts: sss,

		processDir:     pd,
		processScripts: pss,

		finishScripts: fss,
	}
}

type lifecycle struct {
	step LifecycleStep

	config  *lifecycleConfig
	options *LifecycleOptions
}

func newLifecycle(eplcc *lifecycleConfig, eplco *LifecycleOptions) *lifecycle {

	return &lifecycle{
		config:  eplcc,
		options: eplco,
	}
}

func (eplc *lifecycle) exec(ctx context.Context) *Exec {

	log.Trace("lifecycle.exec called")

	return NewExec(ctx).WithTimeout(eplc.options.TerminateAllOnExitTimeout)
}

func (eplc *lifecycle) run(ctx context.Context) (int, error) {

	log.Trace("lifecycle.Run called")

	runCtx, cancelRun := context.WithCancel(ctx)
	defer eplc.exit(ctx, cancelRun) // exit

	processCtx, cancelProcess := context.WithCancel(runCtx)

	// catch first interrupt
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {

		<-interrupt

		log.Info("Container execution aborted (SIGINT, SIGTERM or SIGQUIT signal received)")

		if eplc.step == LifecycleStepProcess {
			cancelProcess()
		} else {
			cancelRun()
		}

	}()

	exitCode := 0
	var err error

	// startup
	if !eplc.options.SkipStartup {
		exitCode, err = eplc.runStep(runCtx, LifecycleStepStartup, eplc.options.PreStartupCmds, eplc.startup)
		if err != nil {
			return 1, err
		}
	}

	// process
	if !eplc.options.SkipProcess {
		exitCode, err = eplc.runStep(processCtx, LifecycleStepProcess, eplc.options.PreProcessCmds, eplc.process)
		if err != nil {
			return 1, err
		}
	}

	// finish
	if !eplc.options.SkipFinish {
		exitCode, err = eplc.runStep(runCtx, LifecycleStepFinish, eplc.options.PreFinishCmds, eplc.finish)
		if err != nil {
			return 1, err
		}
	}

	return exitCode, nil
}

func (eplc *lifecycle) runStep(ctx context.Context, step LifecycleStep, preCmds []string, stepFunc func(context.Context) error) (int, error) {

	log.Trace("lifecycle.runStep called")

	if ctx.Err() == context.Canceled {
		log.Debugf("Ignoring %v lifecycle step (container execution aborted) ...", step)
		return 0, nil
	}

	eplc.step = step

	log.Debugf("Starting %v lifecycle step ...", step)
	if err := eplc.execPreCommands(ctx, preCmds, eplc.step); err != nil {
		return 1, err
	}

	exitCode := 0

	if err := stepFunc(ctx); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			exitCode = exiterr.ExitCode()
			log.Errorf("Error during %v lifecycle: %v", step, err.Error())
		} else if !errors.Is(err, context.Canceled) {
			return 1, err
		}
	}

	return exitCode, nil
}

func (eplc *lifecycle) startup(ctx context.Context) error {

	log.Trace("lifecycle.startup called")

	if err := eplc.exec(ctx).Scripts(eplc.config.startupScripts); err != nil {
		return err
	}

	return nil
}

func (eplc *lifecycle) process(ctx context.Context) error {

	log.Trace("lifecycle.process called")

	nbProcesses := len(eplc.config.processScripts)

	if nbProcesses == 1 {
		return eplc.execSingleProcess(ctx)
	}

	if nbProcesses > 1 {
		return eplc.execMultipleProcesses(ctx)
	}

	return eplc.execNoProcess(ctx)
}

func (eplc *lifecycle) finish(ctx context.Context) error {

	log.Tracef("lifecycle.finish called")

	if err := eplc.exec(ctx).Scripts(eplc.config.finishScripts); err != nil {
		return err
	}

	return nil
}

func (eplc *lifecycle) exit(ctx context.Context, ctxCancelFunc context.CancelFunc) {

	log.Tracef("lifecycle.exit called with ctxCancelFunc: %v", ctxCancelFunc)

	eplc.step = LifecycleStepExit

	log.Trace("Calling ctxCancelFunc ...")
	ctxCancelFunc()

	if eplc.options.TerminateAllOnExit {
		eplc.killAll()
	}

	log.Debug("Starting exit lifecycle step ...")
	if err := eplc.execPreCommands(ctx, eplc.options.PreExitCmds, eplc.step); err != nil {
		log.Debug(err.Error())
	}

	log.Info("Exiting ...")
}

func (eplc *lifecycle) killAll() {

	log.Trace("lifecycle.killAll called")

	// if no others proccess is running return
	if pids, _ := ListPids(); len(pids) == 0 {
		return
	}

	// security to not kill all processes if container-baseimage is run outside a container
	if os.Getpid() != 1 {
		log.Warning("Current process is not PID 1: ignoring terminating all processes ...")
		return
	}

	timeout := eplc.options.TerminateAllOnExitTimeout
	log.Infof("Terminating all processes (timeout: %v) ...", timeout)

	if err := KillAll(syscall.SIGTERM); err != nil {
		log.Errorf("Error terminating all processes: %v", err.Error())
	}

	timer := time.AfterFunc(timeout, func() {
		log.Info("Terminating all processes: timeout reached, killing all processes ...")
		if err := KillAll(syscall.SIGKILL); err != nil {
			log.Errorf("Error killing all processes: %v", err.Error())
		}
	})
	defer timer.Stop()

	for {
		pids, _ := ListPids()
		childs := len(pids)
		if childs == 0 {
			break
		}
		log.Debugf("%v child processes still running ...", childs)
		log.Tracef("child processes: %v ...", pids)
		time.Sleep(250 * time.Millisecond)
	}
}

func (eplc *lifecycle) execPreCommands(ctx context.Context, commands []string, epls LifecycleStep) error {

	log.Tracef("lifecycle.execPreCommand called with commands: %v, lifecycleStep: %v", commands, epls)

	for _, command := range commands {
		log.Infof("Running pre-%v command %v ...", epls, command)
		if err := eplc.exec(ctx).Shlex(command); err != nil {
			return err
		}
	}

	return nil
}

func (eplc *lifecycle) execNoProcess(ctx context.Context) error {

	log.Trace("lifecycle.execNoProcess")

	g, subCtx := errgroup.WithContext(ctx)

	if len(eplc.options.Commands) == 0 {
		eplc.options.RunBash = true
	} else {
		g.Go(func() error {
			return eplc.exec(subCtx).Command(eplc.options.Commands...)
		})
	}

	if eplc.options.RunBash {
		g.Go(func() error {
			return eplc.exec(subCtx).Command("bash")
		})
	}

	return g.Wait()
}

func (eplc *lifecycle) execSingleProcess(ctx context.Context) error {

	log.Trace("lifecycle.execSingleProcess called")

	g, subCtx := errgroup.WithContext(ctx)

	g.Go(func() error {

		process := eplc.config.processScripts[0]
		if len(eplc.options.Commands) > 0 {
			process = fmt.Sprintf("%v %v", process, strings.Join(eplc.options.Commands, " "))
		}

		return eplc.exec(subCtx).Script(process)
	})

	if eplc.options.RunBash {
		g.Go(func() error {
			return eplc.exec(subCtx).Command("bash")
		})
	}

	return g.Wait()
}

func (eplc *lifecycle) execMultipleProcesses(ctx context.Context) error {

	log.Trace("lifecycle.execMultipleProcesses called")

	g, subCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		log.Info("Running multiprocess container ...")
		return eplc.exec(subCtx).Command("runsvdir", "-P", eplc.config.processDir)
	})

	if !eplc.options.RestartProcesses {
		g.Go(func() error {

			nbProcesses := len(eplc.config.processScripts)
			log.Debugf("Waiting %v services to start ...", nbProcesses)
			for {
				stdout := &bytes.Buffer{}
				if err := eplc.exec(subCtx).WithStdout(stdout).Shlex(fmt.Sprintf("bash -c \"sv status %v/* | grep -c -v '^fail:'\"", eplc.config.processDir)); err != nil {
					log.Trace(err.Error())
				}

				startedProccesses, err := strconv.Atoi(strings.TrimSuffix(stdout.String(), "\n"))
				if err != nil {
					return err
				}

				log.Debugf("Started services %v/%v ...", startedProccesses, nbProcesses)
				if startedProccesses == nbProcesses {
					break
				}

				time.Sleep(250 * time.Millisecond)
			}

			log.Info("Modifying settings to not restart terminated services ...")
			return eplc.exec(subCtx).Shlex(fmt.Sprintf("bash -c \"sv once %v/*\"", eplc.config.processDir))
		})
	}

	if len(eplc.options.Commands) > 0 {
		g.Go(func() error {
			return eplc.exec(subCtx).Command(eplc.options.Commands...)
		})
	}

	if eplc.options.RunBash {
		g.Go(func() error {
			return eplc.exec(subCtx).Command("bash")
		})
	}

	if err := g.Wait(); err != nil && err != context.Canceled {
		log.Warning(err.Error())
	}

	timeout := eplc.options.TerminateAllOnExitTimeout
	log.Infof("Terminating all services (timeout: %v) ...", timeout)

	return eplc.exec(context.Background()).Shlex(fmt.Sprintf("bash -c 'sv -w %d force-shutdown %v/*'", timeout/time.Second, eplc.config.processDir))
}

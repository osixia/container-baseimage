package core_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/osixia/container-baseimage/config"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/helpers"
)

func setup(t *testing.T) core.Core {

	t.Helper()
	dir := t.TempDir()
	c := *config.CoreConfig
	c.FilesystemConfig = &core.FilesystemConfig{RootPath: filepath.Join(dir, "container"), RunRootPath: filepath.Join(dir, "run"), EnvironmentFilesPrefix: ".env"}
	core.CoreConfig = &c

	if err := core.Init(&c); err != nil {
		t.Fatal(err)
	}

	ci := core.Instance()

	if err := ci.Filesystem().Create(); err != nil {
		t.Fatal(err)
	}

	return ci
}

func service(t *testing.T, c core.Core, name, script string) core.Service {

	t.Helper()
	dir := filepath.Join(c.Filesystem().Paths().Services, name)

	if err := os.MkdirAll(filepath.Join(dir, ".tags"), 0755); err != nil {
		t.Fatal(err)
	}

	if err := helpers.CreateMarker(filepath.Join(dir, ".tags", name)); err != nil {
		t.Fatal(err)
	}

	if script != "" {
		if err := os.WriteFile(filepath.Join(dir, "process.sh"), []byte("#!/bin/sh\n"+script), 0755); err != nil {
			t.Fatal(err)
		}
	}

	s, err := c.Services().Get(name)

	if err != nil {
		t.Fatal(err)
	}

	return s
}

func TestServiceTagExclusions(t *testing.T) {

	c := setup(t)
	service(t, c, "web", "")
	service(t, c, "db", "")
	ss, err := c.Services().List(core.WithoutServicesNames([]string{"tag:web"}), core.HandleServicesTagPrefixInNames("tag:"))

	if err != nil || len(ss) != 1 || ss[0].Name() != "db" {
		t.Fatalf("services=%v err=%v", ss, err)
	}

	ss, err = c.Services().List(core.WithServicesNames([]string{"web", "tag:db"}), core.WithoutServicesNames([]string{"tag:web"}), core.HandleServicesTagPrefixInNames("tag:"))

	if err != nil || len(ss) != 1 || ss[0].Name() != "db" {
		t.Fatalf("mixed selection=%v err=%v", ss, err)
	}

	ss, err = c.Services().List(core.WithServicesNames([]string{"tag:missing"}), core.HandleServicesTagPrefixInNames("tag:"))

	if err != nil || len(ss) != 0 {
		t.Fatalf("unmatched selection=%v err=%v", ss, err)
	}
}

func TestProcessTagsAndPID(t *testing.T) {

	c := setup(t)
	s := service(t, c, "web", "")
	p, err := c.Processes().New(s)

	if err != nil {
		t.Fatal(err)
	}

	for _, pid := range []string{"999999999", "0", "-1", "invalid", ""} {
		if err := os.WriteFile(p.PIDFile(), []byte(pid), 0644); err != nil {
			t.Fatal(err)
		}

		if p.IsUp() {
			t.Errorf("PID %q reported up", pid)
		}
	}

	if err := os.WriteFile(p.PIDFile(), []byte(fmt.Sprint(os.Getpid())), 0644); err != nil {
		t.Fatal(err)
	}

	if !p.IsUp() {
		t.Fatal("live PID reported down")
	}

	ps, err := c.Processes().List(core.WithProcessesNames([]string{"tag:web"}), core.HandleProcessesTagPrefixInNames("tag:"))

	if err != nil || len(ps) != 1 || ps[0].Name() != "web" {
		t.Fatalf("processes=%v error=%v", ps, err)
	}
}

func TestRejectTraversal(t *testing.T) {

	c := setup(t)

	if _, err := c.Services().Get(".."); err == nil {
		t.Fatal("accepted service traversal")
	}

	if _, err := c.Processes().List(core.WithProcessesNames([]string{"../web"})); err == nil {
		t.Fatal("accepted process traversal")
	}

	if _, err := c.Processes().Get("../web"); err == nil {
		t.Fatal("accepted process construction traversal")
	}

	for _, o := range []*core.GenerateServicesOptions{{Names: []string{"../web"}}, {Names: []string{"web"}, Tags: []string{"../../outside"}}} {
		if _, err := c.Generator().GenerateServices(o); err == nil {
			t.Fatal("accepted generator traversal")
		}
	}
}

func TestLifecyclePreservesExitStatusAndFinish(t *testing.T) {

	c := setup(t)
	s := service(t, c, "web", "exit 7\n")
	marker := filepath.Join(c.Filesystem().Paths().RunRoot, "finished")

	if err := os.WriteFile(filepath.Join(c.Filesystem().Paths().Services, "web", "finish.sh"), []byte("#!/bin/sh\ntouch "+helpers.EscapeShell(marker)+"\n"), 0755); err != nil {
		t.Fatal(err)
	}

	code, err := c.Entrypoint().Run(context.Background(), core.EntrypointOptions{SkipEnvFiles: true, Services: []core.Service{s}})

	if err != nil || code != 7 {
		t.Fatalf("code=%d err=%v", code, err)
	}

	if _, err := os.Stat(marker); err != nil {
		t.Fatal("finish did not run", err)
	}
}

func TestLifecycleRestartBackoff(t *testing.T) {

	c := setup(t)
	marker := filepath.Join(c.Filesystem().Paths().RunRoot, "attempts")
	s := service(t, c, "web", "echo attempt >> "+helpers.EscapeShell(marker)+"\nexit 1\n")
	ctx, cancel := context.WithTimeout(context.Background(), 450*time.Millisecond)
	defer cancel()
	restart := true
	_, err := c.Entrypoint().Run(ctx, core.EntrypointOptions{SkipEnvFiles: true, Services: []core.Service{s}, LifecycleOptions: core.LifecycleOptions{RestartProcesses: &restart}})

	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(marker)

	if err != nil {
		t.Fatal(err)
	}

	count := 0

	for _, b := range data {
		if b == '\n' {
			count++
		}
	}

	if count < 2 || count > 4 {
		t.Fatalf("expected bounded retries, got %d", count)
	}
}

func TestLifecycleParentCancellationReapsOrphans(t *testing.T) {

	c := setup(t)
	ready := filepath.Join(c.Filesystem().Paths().RunRoot, "ready")
	orphan := filepath.Join(c.Filesystem().Paths().RunRoot, "orphan")
	script := "sh -c 'sleep 30 & echo $! > " + helpers.EscapeShell(orphan) + "'\n" +
		"touch " + helpers.EscapeShell(ready) + "\nexec sleep 30\n"
	s := service(t, c, "web", script)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		_, err := c.Entrypoint().Run(ctx, core.EntrypointOptions{
			SkipEnvFiles: true,
			Services:     []core.Service{s},
			LifecycleOptions: core.LifecycleOptions{
				KillAllOnExit:        true,
				KillAllOnExitTimeout: 100 * time.Millisecond,
			},
		})
		done <- err
	}()

	deadline := time.Now().Add(3 * time.Second)
	for {
		if _, err := os.Stat(ready); err == nil {
			break
		}

		if time.Now().After(deadline) {
			cancel()
			t.Fatal("service did not become ready")
		}

		time.Sleep(time.Millisecond)
	}

	data, err := os.ReadFile(orphan)
	if err != nil {
		t.Fatal(err)
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		t.Fatal(err)
	}

	// Outside PID 1, cleanup belongs to the surrounding init rather than us.
	if os.Getpid() != 1 {
		defer syscall.Kill(pid, syscall.SIGKILL)
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("cleanup did not finish after parent cancellation")
	}

	if os.Getpid() == 1 && syscall.Kill(pid, 0) != syscall.ESRCH {
		t.Fatal("orphan remained after cleanup")
	}
}

func TestLifecycleNoRestartDelay(t *testing.T) {

	c := setup(t)
	attempts := filepath.Join(c.Filesystem().Paths().RunRoot, "attempts")
	ready := filepath.Join(c.Filesystem().Paths().RunRoot, "ready")
	script := "echo attempt >> " + helpers.EscapeShell(attempts) + "\n" +
		"if [ \"$(wc -l < " + helpers.EscapeShell(attempts) + ")\" -ge 6 ]; then\n" +
		"  touch " + helpers.EscapeShell(ready) + "\n  exec sleep 30\nfi\nexit 1\n"
	s := service(t, c, "web", script)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	restart := true
	done := make(chan error, 1)
	go func() {
		_, err := c.Entrypoint().Run(ctx, core.EntrypointOptions{
			SkipEnvFiles: true,
			Services:     []core.Service{s},
			LifecycleOptions: core.LifecycleOptions{
				RestartProcesses:     &restart,
				NoRestartDelay:       true,
				KillAllOnExitTimeout: 100 * time.Millisecond,
			},
		})
		done <- err
	}()

	// Six attempts would take at least 3.1 seconds with the default backoff.
	deadline := time.Now().Add(2 * time.Second)
	for {
		if _, err := os.Stat(ready); err == nil {
			break
		}

		if time.Now().After(deadline) {
			cancel()
			<-done
			t.Fatal("flag did not disable the restart delay")
		}

		time.Sleep(time.Millisecond)
	}

	cancel()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

func TestLifecycleNoRestartDelayDoesNotEnableRestart(t *testing.T) {

	c := setup(t)
	attempts := filepath.Join(c.Filesystem().Paths().RunRoot, "attempts")
	s := service(t, c, "web", "echo attempt >> "+helpers.EscapeShell(attempts)+"\nexit 7\n")

	restart := false
	code, err := c.Entrypoint().Run(context.Background(), core.EntrypointOptions{
		SkipEnvFiles:     true,
		Services:         []core.Service{s},
		LifecycleOptions: core.LifecycleOptions{RestartProcesses: &restart, NoRestartDelay: true},
	})
	if err != nil || code != 7 {
		t.Fatalf("code=%d err=%v", code, err)
	}

	data, err := os.ReadFile(attempts)
	if err != nil || string(data) != "attempt\n" {
		t.Fatalf("flag enabled restarting: attempts=%q err=%v", data, err)
	}
}

func TestLifecycleNoRestartDelayHonorsWantedDown(t *testing.T) {

	c := setup(t)
	attempts := filepath.Join(c.Filesystem().Paths().RunRoot, "attempts")
	s := service(t, c, "web", "echo attempt >> "+helpers.EscapeShell(attempts)+"\nexit 1\n")

	p, err := c.Processes().Get("web")
	if err != nil {
		t.Fatal(err)
	}

	if err := p.SetWantedDown(true); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	restart := true
	_, err = c.Entrypoint().Run(ctx, core.EntrypointOptions{
		SkipEnvFiles:     true,
		Services:         []core.Service{s},
		LifecycleOptions: core.LifecycleOptions{RestartProcesses: &restart, NoRestartDelay: true},
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(attempts); !os.IsNotExist(err) {
		t.Fatalf("wanted-down process ran: %v", err)
	}
}

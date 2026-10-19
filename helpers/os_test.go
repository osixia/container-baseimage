package helpers

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Watch
// =============================

func TestWatchParseErrorDoesNotExposeCommand(t *testing.T) {

	const secret = "watch-secret-must-not-appear"
	for _, command := range []string{
		"echo --password " + secret + " '",
		"echo --password " + secret + ` "`,
		"echo --password " + secret + ` \`,
	} {
		t.Run(command[len(command)-1:], func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "watched")
			writeTestFile(t, path, "initial")
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			done := make(chan error, 1)
			go func() {
				done <- Watch(ctx, []string{path}, []string{command}, true)
			}()

			// Keep writing until the watcher is ready and reports the parse error.
			ticker := time.NewTicker(10 * time.Millisecond)
			defer ticker.Stop()
			for {
				select {
				case err := <-done:
					if err == nil || !strings.Contains(err.Error(), "failed to parse") {
						t.Fatalf("expected a command parse error, got %v", err)
					}
					if strings.Contains(err.Error(), secret) {
						t.Fatalf("parse error exposes command arguments: %v", err)
					}
					return
				case <-ticker.C:
					if err := os.WriteFile(path, []byte("changed"), 0644); err != nil {
						t.Fatal(err)
					}
				}
			}
		})
	}
}

func TestWatchProjectedVolume(t *testing.T) {
	root := t.TempDir()
	volume := filepath.Join(root, "volume")
	if err := os.Mkdir(volume, 0755); err != nil {
		t.Fatal(err)
	}
	makeRevision := func(name, content string) {
		t.Helper()
		dir := filepath.Join(volume, name)
		if err := os.Mkdir(dir, 0755); err != nil {
			t.Fatal(err)
		}
		for _, name := range []string{"tls.crt", "tls.key"} {
			writeTestFile(t, filepath.Join(dir, name), content)
		}
	}
	link := func(target, name string) {
		t.Helper()
		if err := os.Symlink(target, filepath.Join(volume, name)); err != nil {
			t.Fatal(err)
		}
	}
	makeRevision("..revision1", "initial\n")
	link("..revision1", "..data")
	link("..data/tls.crt", "tls.crt")
	link("..data/tls.key", "tls.key")
	ready := filepath.Join(volume, "ready")
	writeTestFile(t, ready, "ready")
	result := filepath.Join(root, "result")
	script := filepath.Join(root, "callback.sh")
	// Publish the observed certificate/key pair atomically for the assertions.
	writeTestFile(t, script, "cat \"$1\" \"$2\" > \"$3.tmp\" && mv \"$3.tmp\" \"$3\"\n")
	command := "sh " + EscapeShell(script) + " " + EscapeShell(filepath.Join(volume, "tls.crt")) +
		" " + EscapeShell(filepath.Join(volume, "tls.key")) + " " + EscapeShell(result)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	done := make(chan error, 1)
	go func() {
		done <- Watch(ctx, []string{filepath.Join(volume, "tls.crt"), filepath.Join(volume, "tls.key"), ready}, []string{command}, false)
	}()
	t.Cleanup(func() {
		cancel()
		if err := <-done; err != nil && err != context.Canceled {
			t.Errorf("watch stopped unexpectedly: %v", err)
		}
	})
	waitFor := func(want string, wake bool) {
		t.Helper()
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()
		for {
			if data, err := os.ReadFile(result); err == nil && string(data) == want {
				return
			}
			select {
			case <-ctx.Done():
				t.Fatalf("watch did not observe %q: %v", want, ctx.Err())
			case <-ticker.C:
				if wake {
					writeTestFile(t, ready, "ready")
				}
			}
		}
	}
	// A real callback establishes that the watcher is installed, without relying
	// on a startup sleep before replacing the projected volume's symlink.
	waitFor("initial\ninitial\n", true)
	time.Sleep(50 * time.Millisecond)
	for _, revision := range []string{"..revision2", "..revision3"} {
		makeRevision(revision, revision+"\n")
		link(revision, "..data-next")
		if err := os.Rename(filepath.Join(volume, "..data-next"), filepath.Join(volume, "..data")); err != nil {
			t.Fatal(err)
		}
		waitFor(revision+"\n"+revision+"\n", false)
	}
	// Let any already queued readiness callback finish, then verify that an
	// unrelated sibling event does not invoke the command.
	time.Sleep(50 * time.Millisecond)
	if err := os.Remove(result); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(volume, "unrelated"), "unrelated")
	time.Sleep(100 * time.Millisecond)
	if _, err := os.Stat(result); !os.IsNotExist(err) {
		t.Fatalf("unrelated file triggered the command: %v", err)
	}
}

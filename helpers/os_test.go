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

package helpers

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestListFilesMissingDirectory(t *testing.T) {

	_, err := ListFiles(os.DirFS(filepath.Join(t.TempDir(), "missing")))

	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected missing directory error, got %v", err)
	}
}

func TestEnvsubstTemplatesTrailingSlash(t *testing.T) {

	root := t.TempDir()
	input, output := filepath.Join(root, "input"), filepath.Join(root, "output")

	if err := os.MkdirAll(filepath.Join(input, "nested"), 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(input, "nested", "x.template"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	files, err := EnvsubstTemplates(input+"/", output, ".template")

	if err != nil {
		t.Fatal(err)
	}

	expected := filepath.Join(output, "nested", "x")

	if len(files) != 1 || files[0] != expected {
		t.Fatalf("unexpected outputs: %v", files)
	}

	data, err := os.ReadFile(expected)

	if err != nil || string(data) != "hello" {
		t.Fatalf("output=%q error=%v", data, err)
	}
}

func TestWatchCancellation(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- Watch(ctx, []string{t.TempDir()}, nil, false)
	}()
	cancel()

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("got %v", err)
		}

	case <-time.After(2 * time.Second):
		t.Fatal("watch did not stop")
	}
}

func TestValidateName(t *testing.T) {

	for _, name := range []string{"", ".", "..", "../web", "web/child", "/web", "web\\child", "web\x00"} {
		if ValidateName(name) == nil {
			t.Errorf("accepted %q", name)
		}
	}

	for _, name := range []string{"web", "php-fpm", "tag.name"} {
		if err := ValidateName(name); err != nil {
			t.Error(err)
		}
	}
}

func TestValidateUnixID(t *testing.T) {

	for _, id := range []string{"", "-1", "1 2", "1000 -s /bin/bash", "abc", "1000;id"} {
		if ValidateUnixID(id) == nil {
			t.Errorf("accepted %q", id)
		}
	}

	for _, id := range []string{"0", "1000", "65534"} {
		if err := ValidateUnixID(id); err != nil {
			t.Error(err)
		}
	}
}

func TestValidateUnixName(t *testing.T) {

	for _, name := range []string{"", "-s", "victim -s /bin/bash -G sudo", "a b", "1user", "user;id", "user\n"} {
		if ValidateUnixName(name) == nil {
			t.Errorf("accepted %q", name)
		}
	}

	for _, name := range []string{"nonroot", "_app", "web-1", "Web_2"} {
		if err := ValidateUnixName(name); err != nil {
			t.Error(err)
		}
	}
}

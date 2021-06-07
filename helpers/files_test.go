package helpers

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

// Test helpers
// =============================

func writeTestFile(t *testing.T, path string, data string) {

	t.Helper()
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}
}

func assertFileContent(t *testing.T, path string, want string) {

	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != want {
		t.Fatalf("%v: got %q, want %q", path, data, want)
	}
}

func mkfifo(t *testing.T, path string) {

	t.Helper()
	if err := syscall.Mkfifo(path, 0644); err != nil {
		t.Skipf("mkfifo unavailable: %v", err)
	}
}

// WriteFileAtomic
// =============================

func TestWriteFileAtomicReplacesSymlinkWithoutFollowingIt(t *testing.T) {

	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	writeTestFile(t, target, "outside")
	pidFile := filepath.Join(dir, "service.pid")
	if err := os.Symlink(target, pidFile); err != nil {
		t.Fatal(err)
	}

	if err := WriteFileAtomic(pidFile, []byte("42"), 0644, 0755); err != nil {
		t.Fatal(err)
	}

	assertFileContent(t, target, "outside")
	assertFileContent(t, pidFile, "42")
	if fi, err := os.Lstat(pidFile); err != nil || fi.Mode()&os.ModeSymlink != 0 {
		t.Fatalf("pid file must be a regular file, got %v, %v", fi, err)
	}
}

func TestWriteFileAtomicDoesNotModifyHardLinkTarget(t *testing.T) {

	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	writeTestFile(t, target, "outside")
	pidFile := filepath.Join(dir, "service.pid")
	if err := os.Link(target, pidFile); err != nil {
		t.Skipf("hard links unavailable: %v", err)
	}

	if err := WriteFileAtomic(pidFile, []byte("42"), 0644, 0755); err != nil {
		t.Fatal(err)
	}

	assertFileContent(t, target, "outside")
	assertFileContent(t, pidFile, "42")
}

func TestWriteFileAtomicRefusesSymlinkedParentDirectory(t *testing.T) {

	dir := t.TempDir()
	outside := filepath.Join(dir, "outside")
	if err := os.Mkdir(outside, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(dir, "processes")); err != nil {
		t.Fatal(err)
	}

	if err := WriteFileAtomic(filepath.Join(dir, "processes", "service.pid"), []byte("42"), 0644, 0755); err == nil {
		t.Fatal("expected an error through a symlinked directory")
	}

	entries, err := os.ReadDir(outside)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("nothing must be written through the symlink, found %v", entries)
	}
}

func TestWriteFileAtomicSetsPermissionsAndLeavesNoTemporaryFile(t *testing.T) {

	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "file")

	if err := WriteFileAtomic(path, []byte("data"), 0640, 0755); err != nil {
		t.Fatal(err)
	}

	fi, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode().Perm() != 0640 {
		t.Fatalf("got mode %v, want 0640", fi.Mode().Perm())
	}
	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected only the written file, found %v", entries)
	}
}

// WriteFileInPlace
// =============================

func TestWriteFileInPlaceKeepsInodeAndFollowsSymlink(t *testing.T) {

	dir := t.TempDir()
	target := filepath.Join(dir, "target.conf")
	writeTestFile(t, target, "old")
	link := filepath.Join(dir, "link.conf")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	before, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}

	if err := WriteFileInPlace(link, []byte("new"), 0644, 0755); err != nil {
		t.Fatal(err)
	}

	assertFileContent(t, target, "new")
	after, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	if !os.SameFile(before, after) {
		t.Fatal("user provided destinations must keep their inode (bind mounts)")
	}
}

// ReadFileSafe
// =============================

func TestReadFileSafe(t *testing.T) {

	dir := t.TempDir()
	regular := filepath.Join(dir, "regular")
	writeTestFile(t, regular, "123")
	symlink := filepath.Join(dir, "symlink")
	if err := os.Symlink(regular, symlink); err != nil {
		t.Fatal(err)
	}
	big := filepath.Join(dir, "big")
	writeTestFile(t, big, strings.Repeat("1", 33))

	data, err := ReadFileSafe(regular, 32)
	if err != nil || string(data) != "123" {
		t.Fatalf("got %q, %v", data, err)
	}

	for name, path := range map[string]string{
		"symlink":   symlink,
		"directory": dir,
		"too big":   big,
		"missing":   filepath.Join(dir, "missing"),
	} {
		if _, err := ReadFileSafe(path, 32); err == nil {
			t.Errorf("%v: expected an error", name)
		}
	}
}

func TestReadFileSafeDoesNotBlockOnFIFO(t *testing.T) {

	fifo := filepath.Join(t.TempDir(), "service.pid")
	mkfifo(t, fifo)

	done := make(chan error, 1)
	go func() {
		_, err := ReadFileSafe(fifo, 32)
		done <- err
	}()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected an error on a FIFO")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("ReadFileSafe blocked on a FIFO")
	}
}

func TestIsFileSafe(t *testing.T) {

	dir := t.TempDir()
	regular := filepath.Join(dir, "regular")
	writeTestFile(t, regular, "")
	symlink := filepath.Join(dir, "symlink")
	if err := os.Symlink(regular, symlink); err != nil {
		t.Fatal(err)
	}
	fifo := filepath.Join(dir, "fifo")
	mkfifo(t, fifo)

	if !IsFileSafe(regular) {
		t.Error("regular file must be accepted")
	}
	for _, path := range []string{symlink, fifo, dir, filepath.Join(dir, "missing")} {
		if IsFileSafe(path) {
			t.Errorf("%v must be rejected", path)
		}
	}
}

// Markers
// =============================

func TestCreateMarkerPreservesExistingFile(t *testing.T) {

	marker := filepath.Join(t.TempDir(), "service.down")
	writeTestFile(t, marker, "keep")

	if err := CreateMarker(marker); err != nil {
		t.Fatal(err)
	}

	assertFileContent(t, marker, "keep")
}

func TestCreateMarkerRefusesSymlink(t *testing.T) {

	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	marker := filepath.Join(dir, "service.down")
	if err := os.Symlink(target, marker); err != nil {
		t.Fatal(err)
	}

	if err := CreateMarker(marker); err == nil {
		t.Fatal("expected an error on a symlink")
	}
	if _, err := os.Lstat(target); !os.IsNotExist(err) {
		t.Fatalf("symlink target must not be created: %v", err)
	}
}

func TestRemoveRefusesSymlinkAndKeepsTarget(t *testing.T) {

	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	writeTestFile(t, target, "outside")
	marker := filepath.Join(dir, "service.down")
	if err := os.Symlink(target, marker); err != nil {
		t.Fatal(err)
	}

	if err := Remove(marker); err == nil {
		t.Fatal("expected an error on a symlink")
	}
	assertFileContent(t, target, "outside")

	if err := Remove(target); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(target); !os.IsNotExist(err) {
		t.Fatalf("regular file must be removed: %v", err)
	}
}

// ConfigureRuntimeDirectory
// =============================

func TestConfigureRuntimeDirectoryCreatesSharedDirectory(t *testing.T) {

	dir := t.TempDir()
	if err := os.Chmod(dir, os.ModeSetgid|0755); err != nil {
		t.Fatal(err)
	}
	runtime := filepath.Join(dir, "processes")

	if err := ConfigureRuntimeDirectory(runtime); err != nil {
		t.Fatal(err)
	}

	fi, err := os.Stat(runtime)
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode().Perm() != 0777 {
		t.Fatalf("got mode %v, want 0777 whatever the umask", fi.Mode().Perm())
	}
	// Kubernetes fsGroup: the setgid bit inherited from the parent is kept
	if fi.Mode()&os.ModeSetgid == 0 {
		t.Fatalf("inherited setgid bit must be kept, got mode %v", fi.Mode())
	}
}

func TestConfigureRuntimeDirectoryKeepsExistingPermissions(t *testing.T) {

	runtime := filepath.Join(t.TempDir(), "processes")
	if err := os.Mkdir(runtime, 0750); err != nil {
		t.Fatal(err)
	}

	if err := ConfigureRuntimeDirectory(runtime); err != nil {
		t.Fatal(err)
	}

	fi, err := os.Stat(runtime)
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode().Perm() != 0750 {
		t.Fatalf("existing directory must not be chmodded, got %v", fi.Mode().Perm())
	}
}

func TestConfigureRuntimeDirectoryRefusesSymlink(t *testing.T) {

	dir := t.TempDir()
	outside := filepath.Join(dir, "outside")
	if err := os.Mkdir(outside, 0700); err != nil {
		t.Fatal(err)
	}
	runtime := filepath.Join(dir, "processes")
	if err := os.Symlink(outside, runtime); err != nil {
		t.Fatal(err)
	}

	if err := ConfigureRuntimeDirectory(runtime); err == nil {
		t.Fatal("expected an error on a symlink")
	}

	fi, err := os.Stat(outside)
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode().Perm() != 0700 {
		t.Fatalf("symlink target must keep its permissions, got %v", fi.Mode().Perm())
	}
}

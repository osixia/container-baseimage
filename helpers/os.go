package helpers

import (
	"context"
	goerrors "errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/google/shlex"
	"golang.org/x/sys/unix"

	"github.com/osixia/container-baseimage/errors"
	"github.com/osixia/container-baseimage/log"
)

// Terminal functions
// =============================

func IsTerminal(file *os.File) bool {

	log.Tracef("IsTerminal called with file: %v", file)

	_, err := unix.IoctlGetTermios(int(file.Fd()), unix.TCGETS)
	return err == nil
}

// Filesystem functions
// =============================

type FilePermFunc func(file string) fs.FileMode
type CopyFunc func(path string, dest string) error

// Create exclusively creates a file without following symlinks.
func Create(name string) (*os.File, error) {

	dir, err := openDirectory(filepath.Dir(name), 0755)
	if err != nil {
		return nil, err
	}
	defer closeQuietly(dir)
	fd, err := unix.Openat(int(dir.Fd()), filepath.Base(name), unix.O_WRONLY|unix.O_CREAT|unix.O_EXCL|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0644)
	if err != nil {
		return nil, &os.PathError{Op: "create", Path: name, Err: err}
	}
	return os.NewFile(uintptr(fd), name), nil
}

// CreateMarker preserves an existing regular file.
func CreateMarker(name string) error {

	f, err := Create(name)
	if goerrors.Is(err, unix.EEXIST) {
		dir, openErr := lookupDirectory(filepath.Dir(name))
		if openErr != nil {
			return openErr
		}
		defer closeQuietly(dir)
		return checkDestination(dir, name)
	}
	if err != nil {
		return err
	}
	return f.Close()
}

func Remove(name string) error {

	log.Tracef("Remove called with name: %v", name)
	log.Debugf("Removing %v ...", name)

	dir, err := lookupDirectory(filepath.Dir(name))
	if err != nil {
		return err
	}
	defer closeQuietly(dir)
	if err := checkDestination(dir, name); err != nil {
		return err
	}
	if err := unix.Unlinkat(int(dir.Fd()), filepath.Base(name), 0); err != nil {
		return &os.PathError{Op: "remove", Path: name, Err: err}
	}
	return nil
}

func ListFiles(fsys fs.FS) (files []string, err error) {

	log.Tracef("ListFiles called with fs: %v", fsys)

	err = fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		files = append(files, path)

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

func Copy(path string, dest string) error {

	log.Tracef("Copy called with path: %v, dest: %v", path, dest)

	isDir, err := IsDir(path)
	if err != nil {
		return err
	}

	var copyFunc CopyFunc = CopyFile
	if isDir {
		copyFunc = CopyDir
	}

	return copyFunc(path, dest)
}

func CopyDir(dir string, dest string) error {

	log.Tracef("CopyDir called with dir: %v, dest: %v", dir, dest)

	files, err := ListFiles(os.DirFS(dir))
	if err != nil {
		return err
	}

	for _, f := range files {

		fp := filepath.Join(dir, f)
		dest := filepath.Join(dest, f)

		if err := CopyFile(fp, dest); err != nil {
			return err
		}
	}

	return nil
}

func CopyFile(file string, dest string) error {

	log.Tracef("CopyFile called with file: %v, dest: %v", file, dest)
	log.Debugf("Copying %v to %v ...", file, dest)

	inputInfo, err := os.Stat(file)
	if err != nil {
		return err
	}

	input, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	return WriteFileAtomic(dest, input, inputInfo.Mode().Perm(), 0755)
}

func IsDir(name string) (bool, error) {

	log.Tracef("IsDir called with name: %v", name)

	if fi, err := os.Stat(name); err != nil || !fi.Mode().IsDir() {
		return false, err
	}

	return true, nil
}

func IsFile(name string) (bool, error) {

	log.Tracef("IsFile called with name: %v", name)

	if fi, err := os.Stat(name); err != nil || fi.Mode().IsDir() {
		return false, err
	}

	return true, nil
}

func NewWatcher(paths ...string) (*fsnotify.Watcher, error) {

	log.Tracef("NewWatcher called with paths: %v", paths)

	if len(paths) < 1 {
		return nil, fmt.Errorf("paths: %w", errors.ErrRequired)
	}

	// Create a new watcher.
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	// Add all paths from the commandline.
	for _, p := range paths {
		err = w.Add(p)
		if err != nil {
			_ = w.Close()
			return nil, err
		}
	}

	return w, nil
}

func Watch(ctx context.Context, paths []string, scripts []string, once bool) error {

	log.Tracef("Watch called with paths: %v, scripts: %v, once: %v", paths, scripts, once)
	log.Infof("Watching changes on %s ...", strings.Join(paths, ", "))

	files := make(map[string]os.FileInfo)
	dirs := make(map[string]bool)
	watched := make(map[string]bool)
	var watchPaths []string

	for _, path := range paths {
		path, err := filepath.Abs(path)
		if err != nil {
			return err
		}

		info, err := os.Stat(path)
		if err != nil {
			return err
		}

		watchPath := path
		if info.IsDir() {
			dirs[path] = true
		} else {
			files[path] = info
			watchPath = filepath.Dir(path)
		}

		if !watched[watchPath] {
			watched[watchPath] = true
			watchPaths = append(watchPaths, watchPath)
		}
	}

	w, err := NewWatcher(watchPaths...)
	if err != nil {
		return err
	}
	defer func() {
		if err := w.Close(); err != nil {
			log.Warning(err.Error())
		}
	}()

	for {
		select {

		case <-ctx.Done():

			return ctx.Err()
		case err, ok := <-w.Errors:
			if !ok { // channel was closed
				return nil
			}

			return err

		case e, ok := <-w.Events:
			if !ok { // channel was closed
				return nil
			}

			if e.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename) == 0 {
				continue
			}

			changed := files[e.Name] != nil || dirs[e.Name] || dirs[filepath.Dir(e.Name)]
			for path, previous := range files {
				if filepath.Dir(path) != filepath.Dir(e.Name) {
					continue
				}
				// A sibling symlink can retarget the file without changing its
				// name, as with Kubernetes projected volumes replacing ..data.
				current, err := os.Stat(path)
				if err != nil {
					if os.IsNotExist(err) {
						continue // Keep the last target until the replacement appears.
					}
					return err
				}
				if !os.SameFile(previous, current) {
					changed = true
				}
				files[path] = current
			}
			if !changed {
				continue
			}

			log.Tracef("recieved watch event %+v", e)

			log.Infof("Change detected on %v", e.Name)

			for _, s := range scripts {
				parts, err := shlex.Split(s)
				if err != nil {
					return fmt.Errorf("failed to parse watched command: %w", err)
				}
				if len(parts) == 0 {
					continue
				}
				if err := NewExec(ctx).Command(parts[0], parts[1:]...).Run(); err != nil {
					return err
				}
			}

			if once {
				return nil
			}

		}
	}

}

func EscapeShell(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

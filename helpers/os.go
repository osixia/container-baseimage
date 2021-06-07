package helpers

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"

	"github.com/osixia/container-baseimage/errors"
	"github.com/osixia/container-baseimage/log"
)

// Filesystem functions
// =============================

type FilePermFunc func(file string) fs.FileMode
type CopyFunc func(path string, dest string) error

func Create(name string) (*os.File, error) {

	log.Tracef("Create called with name: %v", name)

	dir := filepath.Dir(name)

	log.Tracef("Create directory %v", dir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	log.Debugf("Create %v", name)
	return os.Create(name)
}

func Remove(name string) error {

	log.Tracef("Remove called with name: %v", name)
	log.Debugf("Removing %v ...", name)

	return os.Remove(name)
}

func ListFiles(fsys fs.FS) (files []string, err error) {

	log.Tracef("ListFiles called with fs: %v", fsys)

	err = fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
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

	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}

	return os.WriteFile(dest, input, inputInfo.Mode().Perm())
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
	log.Infof("Watching changes on %v ...", paths)

	w, err := NewWatcher(paths...)
	if err != nil {
		return err
	}
	defer func() {
		if err := w.Close(); err != nil {
			log.Error(err.Error())
		}
	}()

	for {
		select {

		case _, ok := <-w.Errors:
			if !ok { // channel was closed
				return nil
			}

		case e, ok := <-w.Events:
			if !ok { // channel was closed
				return nil
			}

			log.Tracef("recieved watch event %+v", e)

			if e.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename) != 0 {

				log.Infof("Change detected on %v", e.Name)

				for _, s := range scripts {
					if err := NewExec(ctx).Command(s).Run(); err != nil {
						return err
					}
				}
			}

		}

		if once {
			return nil
		}
	}

}

func EscapeShell(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

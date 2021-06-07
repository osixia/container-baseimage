package helpers

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

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

	if err := os.Remove(name); err != nil {
		return err
	}

	return nil
}

func Symlink(target string, dest string) error {

	log.Tracef("Symlink called with target: %v, dest: %v", target, dest)

	dir := filepath.Dir(dest)

	log.Tracef("Create directory %v", dir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	log.Debugf("Link %v to %v", target, dest)
	if err := os.Symlink(target, dest); err != nil {
		if link, _ := os.Readlink(dest); link != target {
			return err
		}
	}

	return nil
}

func SymlinkAll(target string, dest string) error {

	log.Tracef("SymlinkAll called with target: %v, dest: %v", target, dest)

	isDir, err := IsDir(target)
	if err != nil {
		return err
	}

	if !isDir {
		return Symlink(target, dest)
	}

	files, err := os.ReadDir(target)
	if err != nil {
		return err
	}

	for _, file := range files {
		if err := Symlink(filepath.Join(target, file.Name()), filepath.Join(dest, file.Name())); err != nil {
			return err
		}
	}

	return nil
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

	if err := copyFunc(path, dest); err != nil {
		return err
	}

	return nil
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

	err = os.WriteFile(dest, input, inputInfo.Mode().Perm())
	if err != nil {
		return err
	}

	return nil
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

func NewFSWatcher(paths ...string) (*fsnotify.Watcher, error) {
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
			return nil, err
		}
	}

	return w, nil
}

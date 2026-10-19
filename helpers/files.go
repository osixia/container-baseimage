package helpers

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"

	"github.com/osixia/container-baseimage/log"
)

// openDirectory walks with directory descriptors: symlinks are never followed,
// and renames cannot redirect subsequent operations to a different directory.
// Directory owners and modes are not restricted: runtime directories can be
// volumes (Kubernetes emptyDir, bind mounts) owned by another user.
// Missing directories are created with perm.
func openDirectory(path string, perm fs.FileMode) (*os.File, error) {
	return walkDirectory(path, true, perm)
}

// lookupDirectory is openDirectory for read and remove operations: a missing
// directory is an error, never created (e.g. with root default permissions
// before the supervisor creates the shared runtime directories).
func lookupDirectory(path string) (*os.File, error) {
	return walkDirectory(path, false, 0)
}

func walkDirectory(path string, create bool, perm fs.FileMode) (*os.File, error) {

	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	fd, err := unix.Open("/", unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	dir := os.NewFile(uintptr(fd), "/")
	for _, part := range strings.Split(strings.TrimPrefix(path, "/"), "/") {
		if part == "" {
			continue
		}
		next, err := unix.Openat(int(dir.Fd()), part, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
		if create && errors.Is(err, unix.ENOENT) {
			if err = unix.Mkdirat(int(dir.Fd()), part, uint32(perm.Perm())); err == nil || errors.Is(err, unix.EEXIST) {
				next, err = unix.Openat(int(dir.Fd()), part, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
			}
		}
		closeQuietly(dir)
		if err != nil {
			return nil, fmt.Errorf("opening directory %q: %w", path, err)
		}
		dir = os.NewFile(uintptr(next), part)
	}
	return dir, nil
}

// closeQuietly closes descriptors only used to read or as *at anchors:
// a close error cannot lose written data.
func closeQuietly(f *os.File) {
	_ = f.Close()
}

// ConfigureRuntimeDirectory is used by install and startup. Runtime directories
// are shared by all container users: a created directory is made world writable
// so any user can start a process stopped by another one, keeping inherited bits
// (e.g. Kubernetes fsGroup setgid). An existing directory (image layer, volume,
// tmpfs) is used as is: its permissions belong to whoever created or mounted it.
func ConfigureRuntimeDirectory(path string) error {
	return configureRuntimeDirectory(path, 0777)
}

// ConfigureTemporaryDirectory creates a shared temporary directory with the
// sticky bit. Existing directory permissions and inherited setgid are preserved.
func ConfigureTemporaryDirectory(path string) error {
	return configureRuntimeDirectory(path, 01777)
}

func configureRuntimeDirectory(path string, mode uint32) error {

	parent, err := openDirectory(filepath.Dir(path), 0755)
	if err != nil {
		return err
	}
	defer closeQuietly(parent)
	name := filepath.Base(path)
	// an existing directory may also fail with EROFS or EACCES: try to open it
	mkdirErr := unix.Mkdirat(int(parent.Fd()), name, 0755)
	created := mkdirErr == nil
	fd, err := unix.Openat(int(parent.Fd()), name, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
	if err != nil {
		if mkdirErr != nil && !errors.Is(mkdirErr, unix.EEXIST) {
			err = mkdirErr
		}
		return fmt.Errorf("opening directory %q: %w", path, err)
	}
	dir := os.NewFile(uintptr(fd), path)
	defer closeQuietly(dir)
	if !created {
		log.Tracef("Runtime directory %v already exists, keeping its permissions", path)
		return nil
	}
	var stat unix.Stat_t
	if err := unix.Fstat(fd, &stat); err != nil {
		return &os.PathError{Op: "stat", Path: path, Err: err}
	}
	if err := unix.Fchmod(fd, stat.Mode&07777|mode); err != nil {
		return &os.PathError{Op: "chmod", Path: path, Err: err}
	}
	return nil
}

// checkRegularFile rejects FIFOs, devices and directories. Owners are not checked:
// runtime files are shared by every container user (e.g. processes status or stop
// run with docker exec as any user), at worst another user can disturb the state.
func checkRegularFile(stat *unix.Stat_t, name string) error {

	if stat.Mode&unix.S_IFMT != unix.S_IFREG {
		return fmt.Errorf("file %q must be a regular file", name)
	}
	return nil
}

// checkDestination checks path, an entry of dir, is missing or a regular file.
func checkDestination(dir *os.File, path string) error {

	var stat unix.Stat_t
	err := unix.Fstatat(int(dir.Fd()), filepath.Base(path), &stat, unix.AT_SYMLINK_NOFOLLOW)
	if errors.Is(err, unix.ENOENT) {
		return nil
	}
	if err != nil {
		return &os.PathError{Op: "stat", Path: path, Err: err}
	}
	return checkRegularFile(&stat, path)
}

// WriteFileInPlace follows symlinks and keeps the destination inode, owner and
// mounts (bind-mounted files cannot be replaced by rename). Use it for user provided paths.
func WriteFileInPlace(path string, data []byte, perm fs.FileMode, dirPerm fs.FileMode) error {

	if err := os.MkdirAll(filepath.Dir(path), dirPerm); err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	if _, err = file.Write(data); err == nil {
		err = file.Chmod(perm)
	}
	closeErr := file.Close()
	if err != nil {
		return err
	}
	return closeErr
}

// WriteFileAtomic creates a private temporary file, then publishes it by rename.
// Existing files are never opened for writing: rename replaces the destination
// entry itself, including a planted symlink or hard link, without following it.
func WriteFileAtomic(path string, data []byte, perm fs.FileMode, dirPerm fs.FileMode) error {

	dir, err := openDirectory(filepath.Dir(path), dirPerm)
	if err != nil {
		return err
	}
	defer closeQuietly(dir)
	name := filepath.Base(path)
	tmp := ".container-" + rand.Text()
	fd, err := unix.Openat(int(dir.Fd()), tmp, unix.O_WRONLY|unix.O_CREAT|unix.O_EXCL|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0600)
	if err != nil {
		return &os.PathError{Op: "create", Path: filepath.Join(filepath.Dir(path), tmp), Err: err}
	}
	// Usually already renamed: only cleans up after a failure.
	defer func() { _ = unix.Unlinkat(int(dir.Fd()), tmp, 0) }()
	file := os.NewFile(uintptr(fd), filepath.Join(filepath.Dir(path), tmp))
	if _, err = file.Write(data); err == nil {
		err = file.Chmod(perm.Perm())
	}
	closeErr := file.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	if err := unix.Renameat(int(dir.Fd()), tmp, int(dir.Fd()), name); err != nil {
		return &os.PathError{Op: "rename", Path: path, Err: err}
	}
	return nil
}

// IsFileSafe reports whether path is a regular file, without following
// symlinks nor reading it: markers can be any size.
func IsFileSafe(path string) bool {

	dir, err := lookupDirectory(filepath.Dir(path))
	if err != nil {
		return false
	}
	defer closeQuietly(dir)
	var stat unix.Stat_t
	if err := unix.Fstatat(int(dir.Fd()), filepath.Base(path), &stat, unix.AT_SYMLINK_NOFOLLOW); err != nil {
		return false
	}
	return checkRegularFile(&stat, path) == nil
}

// ReadFileSafe reads at most maxSize bytes: runtime files can be replaced
// by any container user, a huge file must not exhaust the supervisor memory.
func ReadFileSafe(path string, maxSize int64) ([]byte, error) {

	dir, err := lookupDirectory(filepath.Dir(path))
	if err != nil {
		return nil, err
	}
	defer closeQuietly(dir)
	fd, err := unix.Openat(int(dir.Fd()), filepath.Base(path), unix.O_RDONLY|unix.O_NOFOLLOW|unix.O_NONBLOCK|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, &os.PathError{Op: "open", Path: path, Err: err}
	}
	file := os.NewFile(uintptr(fd), path)
	defer closeQuietly(file)
	var stat unix.Stat_t
	if err = unix.Fstat(fd, &stat); err != nil {
		return nil, &os.PathError{Op: "stat", Path: path, Err: err}
	}
	if err = checkRegularFile(&stat, path); err != nil {
		return nil, err
	}
	data, err := io.ReadAll(io.LimitReader(file, maxSize+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxSize {
		return nil, fmt.Errorf("file %q is larger than %v bytes", path, maxSize)
	}
	return data, nil
}

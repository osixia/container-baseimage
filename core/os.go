package core

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/osixia/container-baseimage/log"
)

type FilePermFunc func(file string) fs.FileMode
type CopyFunc func(path string, dest string) error

func Symlink(target string, dest string) error {

	log.Tracef("Symlink called with target: %v, dest: %v", target, dest)

	log.Tracef("Create directory %v", filepath.Dir(dest))
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
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

func ListPids() ([]int, error) {

	log.Tracef("ListPids called")

	var ret []int

	d, err := os.Open("/proc")
	if err != nil {
		return nil, err
	}
	defer d.Close()

	fnames, err := d.Readdirnames(-1)
	if err != nil {
		return nil, err
	}
	for _, fname := range fnames {
		pid, err := strconv.ParseInt(fname, 10, 32)
		if err != nil {
			// if not numeric name, just skip
			continue
		}

		ipid := int(pid)

		// ignore self pid
		if ipid == os.Getpid() {
			continue
		}

		// ignore zombie process
		if zombie, err := isZombie(ipid); err != nil {
			return nil, err
		} else if zombie {
			continue
		}

		ret = append(ret, ipid)
	}

	return ret, nil
}

func KillAll(sig syscall.Signal) error {

	log.Tracef("KillAll called with sig: %v", sig)

	pids, err := ListPids()
	if err != nil {
		return nil
	}

	log.Tracef("pids: %v", pids)

	for _, pid := range pids {
		log.Tracef("Sending %v to pid %v ...", sig, pid)
		if err := syscall.Kill(pid, sig); err != nil {
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

func isZombie(pid int) (bool, error) {
	// Construire le chemin vers le fichier stat du processus
	statPath := fmt.Sprintf("/proc/%d/stat", pid)

	// Lire le contenu du fichier stat
	data, err := ioutil.ReadFile(statPath)
	if err != nil {
		return false, err
	}

	// Convertir les données en chaîne de caractères et les diviser en champs
	fields := strings.Fields(string(data))

	// Le troisième champ contient l'état du processus
	if len(fields) >= 3 && fields[2] == "Z" {
		return true, nil // Le processus est un zombie
	}

	return false, nil // Le processus n'est pas un zombie
}

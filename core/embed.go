package core

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/osixia/container-baseimage/log"
)

func CopyEmbedDir(efs *embed.FS, dest string, filePermFunc FilePermFunc) error {

	log.Tracef("CopyEmbedDir called with efs: %+v, dest: %v, filePermFunc: %+v", efs, dest, filePermFunc)

	files, err := ListFiles(efs)
	if err != nil {
		return err
	}

	for _, f := range files {
		// remove first directory from file path
		fp := filepath.Join(strings.Split(f, "/")[1:]...)

		// append dest path to file path
		fp = filepath.Join(dest, fp)

		perm := filePermFunc(fp)

		if err := CopyEmbedFile(efs, f, fp, perm); err != nil {
			return err
		}
	}

	return nil
}

func CopyEmbedFile(efs *embed.FS, file string, dest string, perm fs.FileMode) error {

	log.Tracef("CopyEmbedFile called with efs: %+v, file: %v, dest: %v, perm: %v", efs, file, dest, perm)

	log.Debugf("Copying %v to %v ...", file, dest)
	fc, err := efs.ReadFile(file)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}

	if err := os.WriteFile(dest, fc, perm); err != nil {
		return err
	}

	return nil
}

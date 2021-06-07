package helpers

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/a8m/envsubst"

	"github.com/osixia/container-baseimage/log"
)

// Envsubst functions
// =============================

func Envsubst(s string) (string, error) {

	log.Tracef("Envsubst called with s: %v", s)

	log.Debugf("Running envsubst on %v ...", s)
	return envsubst.String(s)
}

type writeFileFunc func(path string, data []byte, perm fs.FileMode, dirPerm fs.FileMode) error

// EnvsubstFile writes output in place: it may be a user provided or bind-mounted file.
func EnvsubstFile(input string, output string) error {
	return envsubstFile(input, output, WriteFileInPlace)
}

func envsubstFile(input string, output string, writeFile writeFileFunc) error {

	log.Tracef("EnvsubstFile called with input: %v, output: %v", input, output)

	inputInfo, err := os.Stat(input)
	if err != nil {
		return err
	}

	inputDir := filepath.Dir(input)
	inputDirInfo, err := os.Stat(inputDir)
	if err != nil {
		return err
	}

	log.Debugf("Running envsubst on input %v ...", input)
	bytes, err := envsubst.ReadFile(input)
	if err != nil {
		return err
	}

	if _, err := os.Stat(output); err == nil {
		log.Warningf("File %v already exists and will be overwritten", output)
	}

	// only copy permission bits: never setuid, setgid or sticky
	return writeFile(output, bytes, inputInfo.Mode().Perm(), inputDirInfo.Mode().Perm())
}

func EnvsubstTemplates(templatesDir string, outputDir string, filesSuffix string) ([]string, error) {
	return envsubstTemplates(templatesDir, outputDir, filesSuffix, WriteFileInPlace)
}

// EnvsubstTemplatesAtomic never follows symlinks: use it for outputs in shared runtime directories.
func EnvsubstTemplatesAtomic(templatesDir string, outputDir string, filesSuffix string) ([]string, error) {
	return envsubstTemplates(templatesDir, outputDir, filesSuffix, WriteFileAtomic)
}

func envsubstTemplates(templatesDir string, outputDir string, filesSuffix string, writeFile writeFileFunc) ([]string, error) {

	log.Tracef("EnvsubstTemplates called with templatesDir: %v, outputDir: %v, filesSuffix: %v", templatesDir, outputDir, filesSuffix)

	log.Debugf("EnvsubstTemplates %v files from %v to %v ...", filesSuffix, templatesDir, outputDir)
	log.Debugf("Environment variables:\n%v", strings.Join(os.Environ(), "\n"))

	var files []string
	err := filepath.Walk(templatesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, filesSuffix) {
			relativePath, err := filepath.Rel(templatesDir, path)
			if err != nil {
				return err
			}

			outputPath := filepath.Join(outputDir, strings.TrimSuffix(relativePath, filesSuffix))

			log.Infof("Running envsubst on %v output to %v ...", path, outputPath)
			if err := envsubstFile(path, outputPath, writeFile); err != nil {
				return err
			}

			files = append(files, outputPath)
		}

		return nil
	})

	return files, err
}

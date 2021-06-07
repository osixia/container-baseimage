package core

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/a8m/envsubst"

	"github.com/osixia/container-baseimage/log"
)

func EnvsubstTemplates(templatesDir string, outputDir string, templatesFilesSuffix string) ([]string, error) {

	log.Tracef("filesystem.EnvsubstTemplates called with templatesDir: %v, outputDir: %v, templatesFilesSuffix: %v", templatesDir, outputDir, templatesFilesSuffix)

	log.Debugf("EnvsubstTemplates %v files from %v to %v ...", templatesFilesSuffix, templatesDir, outputDir)
	log.Debugf("Environment variables:\n%v", strings.Join(os.Environ(), "\n"))

	files := []string{}
	err := filepath.Walk(templatesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, templatesFilesSuffix) {
			outputPath := outputDir + strings.TrimSuffix(strings.TrimPrefix(path, templatesDir), templatesFilesSuffix)

			log.Infof("Running envsubst on %v output to %v ...", path, outputPath)
			if err := Envsubst(path, outputPath); err != nil {
				return err
			}

			files = append(files, outputPath)
		}

		return nil
	})

	return files, err
}

func Envsubst(input string, output string) error {

	log.Tracef("filesystem.envsubst called with input: %v, output: %v", input, output)

	inputInfo, err := os.Stat(input)
	if err != nil {
		return err
	}

	inputDir := filepath.Dir(input)
	inputDirInfo, err := os.Stat(inputDir)
	if err != nil {
		return err
	}

	outputDir := filepath.Dir(output)

	log.Debugf("Creating output directory %v ...", outputDir)
	if err := os.MkdirAll(outputDir, inputDirInfo.Mode().Perm()); err != nil {
		return err
	}

	log.Debugf("Running envsubst on input %v ...", input)
	bytes, err := envsubst.ReadFile(input)
	if err != nil {
		return err
	}

	if _, err := os.Stat(output); err == nil {
		log.Warningf("File %v already exists and will be overwrited", output)
	}

	log.Debugf("Creating output file %v ...", output)
	f, err := os.Create(output)
	if err != nil {
		return err
	}
	defer f.Close()

	log.Debugf("Writting output file %v ...", output)
	if _, err := f.Write(bytes); err != nil {
		return err
	}

	log.Debugf("Applying input file %v permissions to %v ...", inputInfo.Mode(), output)
	if err := os.Chmod(output, inputInfo.Mode()); err != nil {
		return err
	}

	return nil
}

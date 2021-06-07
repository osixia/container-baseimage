package cmd

import (
	"github.com/spf13/pflag"

	"github.com/osixia/container-baseimage/build/job"
)

type exportFlags struct {
	buildFlags
}

var exportCmdFlags = &exportFlags{}

var exportCmd = newStepCmd("export", "Run build and export jobs", "e", job.Export, exportCmdFlags)

func init() {
	// flags
	exportCmd.Flags().SortFlags = false
	addExportFlags(exportCmd.Flags(), exportCmdFlags)
}

func addExportFlags(fs *pflag.FlagSet, ef *exportFlags) {
	addBuildFlags(fs, &ef.buildFlags)
}

func (ef *exportFlags) toJobOptions() any {

	return &job.ExportOptions{
		BuildOptions: *ef.buildFlags.toJobOptions().(*job.BuildOptions),
	}
}

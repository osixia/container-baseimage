package job

import (
	"github.com/spf13/pflag"

	"github.com/osixia/container-baseimage/build/helpers"
	"github.com/osixia/container-baseimage/build/job"
)

type ExportFlags struct {
	BuildFlags

	exportDir string
}

var exportCmdFlags = &ExportFlags{}

var exportCmd = helpers.NewJobCmd("export", "Run build and export jobs", "e", job.Export, exportCmdFlags)

func init() {
	// flags
	exportCmd.Flags().SortFlags = false
	AddExportFlags(exportCmd.Flags(), exportCmdFlags)
}

func AddExportFlags(fs *pflag.FlagSet, ef *ExportFlags) {

	fs.StringVarP(&ef.exportDir, "export", "e", "", "export directory")

	AddBuildFlags(fs, &ef.BuildFlags)
}

func (ef *ExportFlags) ToJobOptions() any {

	return &job.ExportOptions{
		BuildOptions: *ef.BuildFlags.ToJobOptions().(*job.BuildOptions),

		ExportDir: ef.exportDir,
	}
}

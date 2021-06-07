package job

import (
	"github.com/spf13/pflag"

	"github.com/osixia/container-baseimage/build/helpers"
	"github.com/osixia/container-baseimage/build/job"
)

type ExportFlags struct {
	BuildFlags

	exportDir  string
	exportFile string

	exportPerArch bool

	exportCompress      bool
	exportCompressImage string
}

var exportCmdFlags = &ExportFlags{}

var exportCmd = helpers.NewJobCmd("export", "Run build and export jobs", "e", job.Export, exportCmdFlags)

func init() {
	// flags
	exportCmd.Flags().SortFlags = false
	AddExportFlags(exportCmd.Flags(), exportCmdFlags)
}

func AddExportFlags(fs *pflag.FlagSet, ef *ExportFlags) {

	fs.StringVar(&ef.exportFile, "export", "", "file to export")
	fs.StringVar(&ef.exportDir, "export-to", "", "directory to export file to\n")

	fs.BoolVar(&ef.exportPerArch, "export-per-arch", false, "export files separately for each architecture\n")

	fs.StringVar(&ef.exportCompressImage, "export-compress-image", "debian:trixie-slim", "container image used to compress exported file")
	fs.BoolVar(&ef.exportCompress, "export-compress", false, "compress exported file\n")

	AddBuildFlags(fs, &ef.BuildFlags)
}

func (ef *ExportFlags) ToJobOptions() any {

	return &job.ExportOptions{
		BuildOptions: *ef.BuildFlags.ToJobOptions().(*job.BuildOptions),

		ExportTo:   ef.exportDir,
		ExportFile: ef.exportFile,

		ExportPerArch: ef.exportPerArch,

		ExportCompress:      ef.exportCompress,
		ExportCompressImage: ef.exportCompressImage,
	}
}

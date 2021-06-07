package job

import (
	"github.com/spf13/pflag"

	"github.com/osixia/container-baseimage/build/helpers"
	"github.com/osixia/container-baseimage/build/job"
)

type TestFlags struct {
	ExportFlags
}

var testCmdFlags = &TestFlags{}

var testCmd = helpers.NewJobCmd("test", "Run build, export and test jobs", "t", job.Test, testCmdFlags)

func init() {
	// flags
	testCmd.Flags().SortFlags = false
	AddTestFlags(testCmd.Flags(), testCmdFlags)
}

func AddTestFlags(fs *pflag.FlagSet, tf *TestFlags) {
	AddExportFlags(fs, &tf.ExportFlags)
}

func (tf *TestFlags) ToJobOptions() any {

	return &job.TestOptions{
		ExportOptions: *tf.ExportFlags.ToJobOptions().(*job.ExportOptions),
	}
}

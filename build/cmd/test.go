package cmd

import (
	"github.com/spf13/pflag"

	"github.com/osixia/container-baseimage/build/job"
)

type testFlags struct {
	exportFlags
}

var testCmdFlags = &testFlags{}

var testCmd = newStepCmd("test", "Run build, export and test jobs", "t", job.Test, testCmdFlags)

func init() {
	// flags
	testCmd.Flags().SortFlags = false
	addTestFlags(testCmd.Flags(), testCmdFlags)
}

func addTestFlags(fs *pflag.FlagSet, tf *testFlags) {
	addExportFlags(fs, &tf.exportFlags)
}

func (tf *testFlags) toJobOptions() any {

	return &job.TestOptions{
		ExportOptions: *tf.exportFlags.toJobOptions().(*job.ExportOptions),
	}
}

package cmd

import (
	"github.com/osixia/container-baseimage/ci/job"
)

type testFlags struct {
	buildFlags
}

var testCmdFlags = &testFlags{}

var testCmd = newStepCmd("test", "Run build and test jobs", "t", job.Test, testCmdFlags)

func init() {
	// flags
	testCmd.Flags().SortFlags = false
	addBuildFlags(testCmd.Flags(), &testCmdFlags.buildFlags)
}

func (tf *testFlags) toJobOptions() interface{} {

	return &job.TestOptions{
		BuildOptions: *tf.buildFlags.toJobOptions().(*job.BuildOptions),
	}
}

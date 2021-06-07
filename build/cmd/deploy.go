package cmd

import (
	"github.com/spf13/pflag"

	"github.com/osixia/container-baseimage/build/job"
)

type deployFlags struct {
	testFlags

	registry string
	username string
	password string

	dryRun bool
}

var deployCmdFlags = &deployFlags{}

var deployCmd = newStepCmd("deploy", "Run build, test and deploy jobs", "d", job.Deploy, deployCmdFlags)

func init() {
	// flags
	deployCmd.Flags().SortFlags = false
	addDeployFlags(deployCmd.Flags(), deployCmdFlags)
	addBuildFlags(deployCmd.Flags(), &deployCmdFlags.buildFlags)
}

func addDeployFlags(fs *pflag.FlagSet, cf *deployFlags) {
	fs.StringVarP(&cf.registry, "registry", "r", "docker.io", "registry")
	fs.StringVarP(&cf.username, "username", "u", "", "username")
	fs.StringVarP(&cf.password, "password", "p", "", "password\n")

	fs.BoolVar(&cf.dryRun, "dry-run", false, "do not deploy images to registry\n")
}

func (df *deployFlags) toJobOptions() interface{} {

	return &job.DeployOptions{
		TestOptions: *df.testFlags.toJobOptions().(*job.TestOptions),

		Registry: df.registry,
		Username: df.username,
		Password: df.password,

		DryRun: df.dryRun,
	}
}

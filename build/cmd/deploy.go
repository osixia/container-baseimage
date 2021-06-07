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

var deployCmd = newStepCmd("deploy", "Run build, export, test and deploy jobs", "d", job.Deploy, deployCmdFlags)

func init() {
	// flags
	deployCmd.Flags().SortFlags = false
	addDeployFlags(deployCmd.Flags(), deployCmdFlags)
}

func addDeployFlags(fs *pflag.FlagSet, df *deployFlags) {
	fs.StringVarP(&df.registry, "registry", "r", "docker.io", "registry")
	fs.StringVarP(&df.username, "username", "u", "", "username")
	fs.StringVarP(&df.password, "password", "p", "", "password\n")

	fs.BoolVar(&df.dryRun, "dry-run", false, "do not deploy images to registry\n")

	addTestFlags(fs, &df.testFlags)
}

func (df *deployFlags) toJobOptions() any {

	return &job.DeployOptions{
		TestOptions: *df.testFlags.toJobOptions().(*job.TestOptions),

		Registry: df.registry,
		Username: df.username,
		Password: df.password,

		DryRun: df.dryRun,
	}
}

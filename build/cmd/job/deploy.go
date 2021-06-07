package job

import (
	"github.com/spf13/pflag"

	"github.com/osixia/container-baseimage/build/helpers"
	"github.com/osixia/container-baseimage/build/job"
)

type DeployFlags struct {
	TestFlags

	registry string
	username string
	password string

	digestsFile string

	dryRun bool
}

var deployCmdFlags = &DeployFlags{}

var deployCmd = helpers.NewJobCmd("deploy", "Run build, export, test and deploy jobs", "d", job.Deploy, deployCmdFlags)

func init() {
	// flags
	deployCmd.Flags().SortFlags = false
	AddDeployFlags(deployCmd.Flags(), deployCmdFlags)
}

func AddDeployFlags(fs *pflag.FlagSet, df *DeployFlags) {
	fs.StringVar(&df.registry, "registry", "docker.io", "registry")
	fs.StringVarP(&df.username, "username", "u", "", "username")
	fs.StringVarP(&df.password, "password", "p", "", "password\n")

	fs.StringVar(&df.digestsFile, "digests-file", "", "append pushed image references (image:tag@digest) to this file\n")

	fs.BoolVar(&df.dryRun, "dry-run", false, "do not deploy images to registry\n")

	AddTestFlags(fs, &df.TestFlags)
}

func (df *DeployFlags) ToJobOptions() any {

	return &job.DeployOptions{
		TestOptions: *df.TestFlags.ToJobOptions().(*job.TestOptions),

		Registry: df.registry,
		Username: df.username,
		Password: df.password,

		DigestsFile: df.digestsFile,

		DryRun: df.dryRun,
	}
}

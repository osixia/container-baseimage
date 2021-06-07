package job

import (
	"github.com/spf13/pflag"

	"github.com/osixia/container-baseimage/build/helpers"
	"github.com/osixia/container-baseimage/build/job"
)

type PublishFlags struct {
	DeployFlags
}

var publishCmdFlags = &PublishFlags{}

var publishCmd = helpers.NewJobCmd("publish", "Run build, export, test, deploy and publish jobs", "p", job.Publish, publishCmdFlags)

func init() {
	// flags
	publishCmd.Flags().SortFlags = false
	AddDeployFlags(publishCmd.Flags(), &publishCmdFlags.DeployFlags)
}

func AddPublishFlags(fs *pflag.FlagSet, tf *PublishFlags) {
	AddDeployFlags(fs, &tf.DeployFlags)
}

func (pf *PublishFlags) ToJobOptions() any {

	return &job.PublishOptions{
		DeployOptions: *pf.DeployFlags.ToJobOptions().(*job.DeployOptions),
	}
}

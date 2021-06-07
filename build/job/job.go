package job

import (
	"context"
	"fmt"
	"os"

	"dagger.io/dagger"
	"github.com/osixia/container-baseimage/build/config"
)

type Job int

const (
	Build  Job = iota
	Export Job = iota
	Test   Job = iota
	Deploy Job = iota
)

type JobFunc func(ctx context.Context, client *dagger.Client, options any) ([]*BuildResult, error)

type BuildTagsFunc func(image *config.Image, options *BuildImageOptions, nonroot bool) []string
type BuildArgsFunc func(options *BuildImageOptions, image *config.Image, platform *config.Platform, tag string) []dagger.BuildArg

var BuildJob JobFunc = DefaultBuildJob
var ExportJob JobFunc = DefaultExportJob
var TestJob JobFunc = DefaultTestJob
var DeployJob JobFunc = DefaultDeployJob

var BuildTags BuildTagsFunc = DefaultBuildTags
var BuildArgs BuildArgsFunc = DefaultBuildArgs

func Run(ctx context.Context, s Job, options any) ([]*BuildResult, error) {

	// dagger client
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := client.Close(); err != nil {
			fmt.Print(err.Error())
		}
	}()

	switch s {
	case Build:
		return BuildJob(ctx, client, options.(*BuildOptions))
	case Export:
		return ExportJob(ctx, client, options.(*ExportOptions))
	case Test:
		return TestJob(ctx, client, options.(*TestOptions))
	case Deploy:
		return DeployJob(ctx, client, options.(*DeployOptions))
	}

	return nil, fmt.Errorf("%v: job unknown", s)
}

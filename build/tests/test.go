package tests

import (
	"context"
	"fmt"
	"path/filepath"

	"dagger.io/dagger"

	"github.com/osixia/dagger-build/job"
)

// Run checks baseimage-specific behavior after the reusable job's version check. CI builds
// each arch on a native runner: run locally with --arches matching the host to avoid emulation.
func Run(ctx context.Context, client *dagger.Client, dockerfile string, builds []*job.BuildResult) error {

	path, err := filepath.Abs(dockerfile)
	if err != nil {
		return fmt.Errorf("dockerfile path: %w", err)
	}
	docs := client.Host().Directory(filepath.Join(filepath.Dir(path), "docs"))

	for _, build := range builds {
		for _, image := range build.Containers {

			if err := Runtime(ctx, image); err != nil {
				return err
			}
			if err := RuntimeControl(ctx, client, image); err != nil {
				return err
			}
			if err := Examples(ctx, image, docs); err != nil {
				return err
			}
			if err := Tutorials(ctx, client, image, docs, build.Image.Distribution); err != nil {
				return err
			}
		}
	}

	return nil
}

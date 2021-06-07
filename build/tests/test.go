package tests

import (
	"context"
	"fmt"
	"path/filepath"

	"dagger.io/dagger"

	"github.com/osixia/dagger-build/job"
)

// Run checks baseimage-specific behavior after the reusable job's version check, only on
// the engine platform: emulated images fail with "exec format error" as Dagger services.
func Run(ctx context.Context, client *dagger.Client, dockerfile string, builds []*job.BuildResult) error {

	path, err := filepath.Abs(dockerfile)
	if err != nil {
		return fmt.Errorf("dockerfile path: %w", err)
	}
	docs := client.Host().Directory(filepath.Join(filepath.Dir(path), "docs"))

	enginePlatform, err := client.DefaultPlatform(ctx)
	if err != nil {
		return fmt.Errorf("engine platform: %w", err)
	}

	tested := 0
	for _, build := range builds {
		for _, image := range build.Containers {

			platform, err := image.Platform(ctx)
			if err != nil {
				return fmt.Errorf("image platform: %w", err)
			}
			if platform != enginePlatform {
				continue
			}
			tested++

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

	if tested == 0 && len(builds) > 0 {
		return fmt.Errorf("no image built for the engine platform %v", enginePlatform)
	}

	return nil
}

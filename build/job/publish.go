package job

import (
	"context"

	"dagger.io/dagger"
)

type PublishOptions struct {
	DeployOptions
}

func DefaultPublish(ctx context.Context, client *dagger.Client, options any) ([]*BuildResult, error) {
	o := options.(*PublishOptions)
	return Deploy(ctx, client, &o.DeployOptions)
}

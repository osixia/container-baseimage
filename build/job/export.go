package job

import (
	"context"

	"dagger.io/dagger"
)

type ExportOptions struct {
	BuildOptions
}

func DefaultExportJob(ctx context.Context, client *dagger.Client, options any) ([]*BuildResult, error) {
	o := options.(*ExportOptions)
	return BuildJob(ctx, client, &o.BuildOptions)
}

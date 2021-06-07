package job

import (
	"context"

	"dagger.io/dagger"
)

type ExportOptions struct {
	BuildOptions

	ExportDir string
}

func DefaultExport(ctx context.Context, client *dagger.Client, options any) ([]*BuildResult, error) {
	o := options.(*ExportOptions)
	return Build(ctx, client, &o.BuildOptions)
}

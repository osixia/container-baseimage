package job

import (
	"context"
	"os"

	"dagger.io/dagger"
)

type Job int

const (
	Build  Job = iota
	Test   Job = iota
	Deploy Job = iota
)

func Run(s Job, ctx context.Context, options interface{}) error {

	// dagger client
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		return err
	}
	defer client.Close()

	switch s {
	case Build:
		_, err = build(ctx, client, options.(*BuildOptions))
	case Test:
		_, err = test(ctx, client, options.(*TestOptions))
	case Deploy:
		_, err = deploy(ctx, client, options.(*DeployOptions))
	}

	return err
}

package job

import (
	"context"
	"fmt"
	"os"

	"dagger.io/dagger"
)

type Job int

const (
	Build  Job = iota
	Test   Job = iota
	Deploy Job = iota
)

func Run(ctx context.Context, s Job, options interface{}) ([]*BuildResult, error) {

	// dagger client
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		return nil, err
	}
	defer client.Close()

	switch s {
	case Build:
		return build(ctx, client, options.(*BuildOptions))
	case Test:
		return test(ctx, client, options.(*TestOptions))
	case Deploy:
		return deploy(ctx, client, options.(*DeployOptions))
	}

	return nil, fmt.Errorf("%v: job unknown", s)
}

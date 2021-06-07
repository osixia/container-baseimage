package cmd

import (
	"testing"
)

func TestNoRestartDelayFlag(t *testing.T) {

	original := cmdFlags.NoRestartDelay
	defer func() {
		cmdFlags.NoRestartDelay = original
	}()

	for _, tc := range []struct {
		name     string
		args     []string
		expected bool
	}{
		{name: "omitted"},
		{name: "bare", args: []string{"--no-restart-delay"}, expected: true},
		{name: "true", args: []string{"--no-restart-delay=true"}, expected: true},
		{name: "false", args: []string{"--no-restart-delay=false"}, expected: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmdFlags.NoRestartDelay = false
			if err := cmd.Flags().Parse(tc.args); err != nil {
				t.Fatal(err)
			}

			options := &flags{
				skipAll: true,
			}
			options.NoRestartDelay = cmdFlags.NoRestartDelay

			entrypoint, err := options.toEntrypointOptions()
			if err != nil {
				t.Fatal(err)
			}

			actual := entrypoint.NoRestartDelay
			if actual != tc.expected {
				t.Fatalf("expected %v, got %v", tc.expected, actual)
			}
		})
	}
}

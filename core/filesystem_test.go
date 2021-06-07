package core

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// unsetTestEnv unsets key for the test and restores it afterwards.
func unsetTestEnv(t *testing.T, key string) {

	t.Helper()
	t.Setenv(key, "")
	if err := os.Unsetenv(key); err != nil {
		t.Fatal(err)
	}
}

func (tc *testCore) addEnvFile(t *testing.T, name string, content string) string {

	t.Helper()
	path := filepath.Join(tc.fs.Paths().Environment, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestListDotEnv(t *testing.T) {

	tc := newTestCore(t)
	b := tc.addEnvFile(t, "b/.env", "")
	a := tc.addEnvFile(t, ".env.a", "")
	tc.addEnvFile(t, "README.md", "")
	tc.addEnvFile(t, "b/env", "")

	files, err := tc.fs.ListDotEnv()
	if err != nil {
		t.Fatal(err)
	}

	if want := []string{a, b}; !slices.Equal(files, want) {
		t.Fatalf("got %v, want %v", files, want)
	}
}

func TestLoadDotEnvKeepsExistingVariables(t *testing.T) {

	tc := newTestCore(t)
	t.Setenv("TEST_EXISTING", "from-env")
	t.Setenv("TEST_EMPTY", "")
	literal := "quotes'\" and ${TEST_NEW}\nsecond line"
	t.Setenv("TEST_LITERAL", literal)
	unsetTestEnv(t, "TEST_NEW")
	unsetTestEnv(t, "TEST_INTERPOLATED")
	file := tc.addEnvFile(t, ".env", "TEST_EXISTING=from-file\nTEST_EMPTY=from-file\nTEST_LITERAL=from-file\nTEST_NEW=new\nTEST_INTERPOLATED=${TEST_NEW}-value\n")

	if err := tc.fs.LoadDotEnv([]string{file}); err != nil {
		t.Fatal(err)
	}

	for key, want := range map[string]string{
		"TEST_EXISTING":     "from-env",
		"TEST_EMPTY":        "",
		"TEST_LITERAL":      literal,
		"TEST_NEW":          "new",
		"TEST_INTERPOLATED": "new-value",
	} {
		if got, ok := os.LookupEnv(key); !ok || got != want {
			t.Errorf("%v: got %q (set: %v), want %q", key, got, ok, want)
		}
	}
}

func TestLoadDotEnvParseErrorDoesNotLeakValues(t *testing.T) {
	for name, content := range map[string]string{
		"missing quotes": "PASSWORD=\"s3cr3t\nBROKEN LINE\n",
		// gotenv includes this entire line in its error without sanitization.
		"invalid line": "PASSWORD s3cr3t\n",
	} {
		t.Run(name, func(t *testing.T) {
			tc := newTestCore(t)
			file := tc.addEnvFile(t, ".env", content)

			err := tc.fs.LoadDotEnv([]string{file})
			if err == nil {
				t.Fatal("expected a syntax error")
			}
			if strings.Contains(err.Error(), "s3cr3t") {
				t.Fatalf("parse errors must not show values: %v", err)
			}
		})
	}
}

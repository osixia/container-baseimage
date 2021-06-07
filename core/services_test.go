package core

import (
	"os"
	"path/filepath"
	"slices"
	"syscall"
	"testing"
)

func TestServicesSortByPriority(t *testing.T) {

	tc := newTestCore(t)
	tc.addService(t, "late", map[string]string{".priority": "600\n"})
	tc.addService(t, "b-default", nil)
	tc.addService(t, "a-default", nil)
	tc.addService(t, "early", map[string]string{".priority": "100"})
	tc.addService(t, "invalid", map[string]string{".priority": "high"})

	ss, err := tc.svcs.List(SortServicesByPriority(true))
	if err != nil {
		t.Fatal(err)
	}

	want := []string{"early", "a-default", "b-default", "invalid", "late"}
	if got := serviceNames(ss); !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestServicesListFilters(t *testing.T) {

	tc := newTestCore(t)
	tc.addService(t, "nginx", map[string]string{".entrypoint": "", ".tags/web": ""})
	tc.addService(t, "php-fpm", map[string]string{".entrypoint": "", ".tags/web": "", ".tags/php": ""})
	tc.addService(t, "cron", map[string]string{".optional": ""})

	for name, tt := range map[string]struct {
		opts []ServicesListOption
		want []string
	}{
		"all":            {nil, []string{"cron", "nginx", "php-fpm"}},
		"linked":         {[]ServicesListOption{WithServicesLinked(true)}, []string{"nginx", "php-fpm"}},
		"not optional":   {[]ServicesListOption{WithServicesOptional(false)}, []string{"nginx", "php-fpm"}},
		"tag":            {[]ServicesListOption{WithServicesTags([]string{"php"})}, []string{"php-fpm"}},
		"excluded tag":   {[]ServicesListOption{WithoutServicesTags([]string{"php"})}, []string{"cron", "nginx"}},
		"excluded name":  {[]ServicesListOption{WithoutServicesNames([]string{"cron"})}, []string{"nginx", "php-fpm"}},
		"empty names":    {[]ServicesListOption{WithServicesNames([]string{})}, []string{}},
		"names and tags": {[]ServicesListOption{WithServicesNames([]string{"cron", "tag:php"}), HandleServicesTagPrefixInNames("tag:")}, []string{"cron", "php-fpm"}},
		"excluded tag in names": {
			[]ServicesListOption{WithServicesNames([]string{"tag:web"}), WithoutServicesNames([]string{"tag:php"}), HandleServicesTagPrefixInNames("tag:")},
			[]string{"nginx"},
		},
	} {
		ss, err := tc.svcs.List(tt.opts...)
		if err != nil {
			t.Errorf("%v: %v", name, err)
			continue
		}
		if got := serviceNames(ss); !slices.Equal(got, tt.want) {
			t.Errorf("%v: got %v, want %v", name, got, tt.want)
		}
	}
}

func TestServicesRejectInvalidOrUnknownNames(t *testing.T) {

	tc := newTestCore(t)
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(tc.fs.Paths().Services, "symlinked")); err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"", "..", "../etc", "a/b", "unknown", "symlinked"} {
		if _, err := tc.svcs.Get(name); err == nil {
			t.Errorf("%q: expected an error", name)
		}
	}

	ss, err := tc.svcs.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(ss) != 0 {
		t.Fatalf("a symlinked service directory must be ignored, got %v", serviceNames(ss))
	}
}

func TestServiceStopSignal(t *testing.T) {

	tc := newTestCore(t)

	for content, want := range map[string]syscall.Signal{
		"quit\n":  syscall.SIGQUIT,
		"SIGHUP":  syscall.SIGHUP,
		" winch ": syscall.SIGWINCH,
		"nope":    0,
	} {
		tc.addService(t, "web", map[string]string{".stop-signal": content})
		s, err := tc.svcs.Get("web")
		if err != nil {
			t.Fatal(err)
		}
		if got := s.StopSignal(); got != want {
			t.Errorf("%q: got %v, want %v", content, got, want)
		}
	}

	tc.addService(t, "default", nil)
	s, err := tc.svcs.Get("default")
	if err != nil {
		t.Fatal(err)
	}
	if got := s.StopSignal(); got != 0 {
		t.Errorf("missing file: got %v, want the default stop signal", got)
	}
}

func TestServiceLinkRefusesSymlinkedMarker(t *testing.T) {

	tc := newTestCore(t)
	dir := tc.addService(t, "web", nil)
	target := filepath.Join(t.TempDir(), "target")
	if err := os.Symlink(target, filepath.Join(dir, ".entrypoint")); err != nil {
		t.Fatal(err)
	}
	s, err := tc.svcs.Get("web")
	if err != nil {
		t.Fatal(err)
	}

	if err := s.Link(); err == nil {
		t.Fatal("expected an error on a symlinked marker")
	}
	if exists(target) {
		t.Fatal("symlink target must not be created")
	}
}

func TestServiceLinkUnlink(t *testing.T) {

	tc := newTestCore(t)
	tc.addService(t, "web", nil)
	s, err := tc.svcs.Get("web")
	if err != nil {
		t.Fatal(err)
	}

	if err := s.Link(); err != nil {
		t.Fatal(err)
	}
	if !s.IsLinked() {
		t.Fatal("service must be linked")
	}
	if err := s.Unlink(); err != nil {
		t.Fatal(err)
	}
	if s.IsLinked() {
		t.Fatal("service must be unlinked")
	}
}

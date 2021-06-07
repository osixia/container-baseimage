package tests

import (
	"context"
	"crypto/rand"
	_ "embed"
	"fmt"
	"strings"
	"time"

	"dagger.io/dagger"

	"github.com/osixia/dagger-build/config"
)

// Examples uses the repository's actual services and environment files on the candidate
// base image. All preparation and test data stay in derived containers.
func Examples(ctx context.Context, image *dagger.Container, docs *dagger.Directory) error {

	user, err := image.User(ctx)
	if err != nil {
		return fmt.Errorf("image user: %w", err)
	}

	for _, name := range []string{"single-process", "multi-process"} {

		example := docs.Directory("examples/" + name)
		fixture := image.WithUser("0").
			WithDirectory("/container/services", example.Directory("services")).
			WithExec([]string{"container", "services", "install"}).
			WithExec([]string{"container", "services", "link"}).
			WithDirectory("/container/environment", example.Directory("environment")).
			WithNewFile("/tmp/baseimage-example-check.sh", exampleCheck, dagger.ContainerWithNewFileOpts{Permissions: 0755}).
			WithUser(user)

		checkCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
		_, err := fixture.WithExec([]string{"sh", "/tmp/baseimage-example-check.sh", name}).Sync(checkCtx)
		cancel()
		if err != nil {
			return fmt.Errorf("%s example: %w", name, err)
		}
	}

	return nil
}

// Tutorials checks the documented web services, which use Debian-style paths and require root
// to install packages and write the web configuration and document root.
func Tutorials(ctx context.Context, client *dagger.Client, image *dagger.Container, docs *dagger.Directory, distribution string) error {

	if distribution != config.Debian && distribution != config.Ubuntu {
		return nil
	}

	user, err := image.User(ctx)
	if err != nil {
		return fmt.Errorf("image user: %w", err)
	}
	if user != "" && user != "0" && user != "root" {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	nginxSource := docs.Directory("tutorials/nginx")
	nginx := image.
		WithExec([]string{"container", "packages", "install", "--update", "--clean", "nginx", "curl"}).
		WithDirectory("/container/services", nginxSource.Directory("services")).
		WithExec([]string{"container", "services", "install"}).
		WithExec([]string{"container", "services", "link"}).
		WithDirectory("/container/environment", nginxSource.Directory("environment")).
		WithEnvVariable("CUSTOM_MESSAGE", "baseimage tutorial runtime check")
	if err := checkTutorialHTTP(ctx, nginx, "/", "baseimage tutorial runtime check"); err != nil {
		return fmt.Errorf("nginx tutorial: %w", err)
	}

	php := nginx.
		WithExec([]string{"container", "packages", "install", "--update", "--clean", "php-fpm"}).
		WithDirectory("/container/services", docs.Directory("tutorials/nginx-php-fpm/services")).
		WithExec([]string{"container", "services", "install"}).
		WithExec([]string{"container", "services", "link"})
	if err := checkTutorialHTTP(ctx, php, "/phpinfo.php", "PHP Version"); err != nil {
		return fmt.Errorf("nginx + php-fpm tutorial: %w", err)
	}

	// The Compose tutorial shares a runtime volume but uses separate PID namespaces.
	state := client.CacheVolume("baseimage-tutorial-split-" + rand.Text())
	split := php.WithMountedCache("/run/container", state, dagger.ContainerWithMountedCacheOpts{Source: php.Directory("/run/container")})
	phpService := split.AsService(dagger.ContainerAsServiceOpts{
		Args: []string{"container", "--exec", "php-fpm", "--kill-all-timeout", "2s"}, NoInit: true,
	})
	if _, err := phpService.Start(ctx); err != nil {
		return fmt.Errorf("split php-fpm tutorial: %w", err)
	}
	defer stopTestService(ctx, phpService)

	nginxService := split.WithExposedPort(80).AsService(dagger.ContainerAsServiceOpts{
		Args: []string{"container", "--exec", "nginx", "--kill-all-timeout", "2s"}, NoInit: true,
	})
	if _, err := nginxService.Start(ctx); err != nil {
		return fmt.Errorf("split nginx tutorial: %w", err)
	}
	defer stopTestService(ctx, nginxService)

	if err := requestTutorialHTTP(ctx, split, nginxService, "/phpinfo.php", "PHP Version"); err != nil {
		return fmt.Errorf("split tutorial HTTP: %w", err)
	}
	if _, err := split.WithUser("10003:10003").WithExec([]string{"container", "processes", "check", "nginx", "php-fpm"}).Sync(ctx); err != nil {
		return fmt.Errorf("split tutorial shared process state: %w", err)
	}

	return nil
}

func checkTutorialHTTP(ctx context.Context, image *dagger.Container, path string, want string) error {

	service := image.WithExposedPort(80).AsService(dagger.ContainerAsServiceOpts{
		Args: []string{"container", "--kill-all-timeout", "2s"}, NoInit: true,
	})
	if _, err := service.Start(ctx); err != nil {
		return fmt.Errorf("start service: %w", err)
	}
	defer stopTestService(ctx, service)

	if err := requestTutorialHTTP(ctx, image, service, path, want); err != nil {
		return err
	}

	if _, err := service.Stop(ctx); err != nil {
		return fmt.Errorf("stop service: %w", err)
	}

	return nil
}

func requestTutorialHTTP(ctx context.Context, image *dagger.Container, service *dagger.Service, path string, want string) error {

	response := image.WithServiceBinding("tutorial", service).
		WithExec([]string{"curl", "--fail", "--silent", "--show-error", "--retry", "20", "--retry-all-errors", "--retry-delay", "1", "--max-time", "2", "http://tutorial" + path}, dagger.ContainerWithExecOpts{RedirectStdout: "/tmp/baseimage-http-response"})
	output, err := response.File("/tmp/baseimage-http-response").Contents(ctx)
	if err != nil {
		return fmt.Errorf("HTTP request to %s: %w", path, err)
	}

	if !strings.Contains(output, want) {
		return fmt.Errorf("HTTP response from %s does not contain %q", path, want)
	}

	return nil
}

//go:embed testdata/example.sh
var exampleCheck string

# Nginx + PHP-FPM Tutorial

This tutorial extends the `nginx` tutorial and turns it into a two-service image.

It demonstrates:

- reusing a previously built image as a base
- adding a second managed service
- running Nginx and PHP-FPM as separate processes
- connecting services through a Unix socket in `/run/container`
- packaging both services in one image while still being able to run them in separate containers

## Prerequisite

This Dockerfile starts from:

```Dockerfile
FROM osixia/nginx-example
```

Build the Nginx tutorial first:

```bash
docker build -t osixia/nginx-example ../nginx
```

## Files

### `Dockerfile`

The image:

- starts from `osixia/nginx-example`
- installs the `php-fpm` package during build
- copies the `php-fpm` service into `/container/services`
- runs `container services install && container services link`

It reuses the Nginx service already present in the parent image.

### `services/php-fpm/install.sh`

Executed at build time. It:

- disables `expose_php`
- sets `cgi.fix_pathinfo=0`
- configures PHP-FPM to listen on `/run/container/php-fpm.sock`
- configures the socket owner and group as `www-data`
- replaces the default Nginx site with the provided PHP-enabled vhost
- creates `/var/www/html/phpinfo.php`

This is the integration point between the two services.

### `services/php-fpm/config/default`

This Nginx vhost:

- serves files from `/var/www/html`
- keeps standard static file handling
- forwards `*.php` requests to `unix:/run/container/php-fpm.sock`

### `services/php-fpm/process.sh`

Executed as the managed PHP-FPM process. It:

- logs the container IP address and the `phpinfo.php` URL
- starts PHP-FPM in the foreground with `--nodaemonize`
- forwards any extra entrypoint arguments to PHP-FPM

### `docker-compose.yml`

This example runs the same multi-process image as two single-purpose containers:

- `nginx` runs with `--exec nginx`
- `php-fpm` runs with `--exec php-fpm`
- both containers mount the same `tmpfs` volume on `/run/container`

This keeps the original Unix socket wiring intact across two containers:

- PHP-FPM still creates `/run/container/php-fpm.sock`
- Nginx still forwards `*.php` requests to that socket
- the shared `tmpfs` keeps socket files and other transient runtime state in memory

## Build

From this directory:

```bash
docker build -t osixia/nginx-php-fpm-example .
```

## Run

```bash
docker run --rm -p 8080:80 osixia/nginx-php-fpm-example
```

Then open:

```text
http://localhost:8080/phpinfo.php
```

Expected runtime behavior:

- the inherited Nginx service starts
- the PHP-FPM service starts in parallel
- Nginx forwards PHP requests to the Unix socket created by PHP-FPM

## Run as Separate Containers

The image is built as a multi-process image for convenience, but it can still be executed one service at a time.

From the tutorial directory:

```bash
docker compose up
```

This compose setup:

- uses the same `osixia/nginx-php-fpm-example` image for both services
- runs `nginx` and `php-fpm` in separate containers with `--exec`
- exposes Nginx on `http://127.0.0.1:8080`
- mounts the same `tmpfs` volume on `/run/container` in both containers
- keeps the deployment closer to common production layouts

Then open:

```text
http://127.0.0.1:8080/
http://127.0.0.1:8080/phpinfo.php
```

## Why Build Multi-Process but Run Separately

This pattern gives you a pragmatic compromise.

Build-time advantages:

- one image can bundle the full application stack
- local development and smoke testing stay simple
- service integration stays close to the code that owns it
- shared setup logic can live in one place

Runtime advantages:

- each service gets its own container lifecycle
- logs are separated by service, which improves observability
- health checks and restart policies can be tuned per service
- resource limits can be set independently for Nginx and PHP-FPM
- scaling becomes more flexible, especially for PHP-FPM workers
- failures are better isolated
- debugging is easier because each container has a single main responsibility
- the topology is closer to what most production platforms expect

Why this compose example keeps the Unix socket:

- the tutorial image configures PHP-FPM to use `/run/container/php-fpm.sock`
- keeping the same socket in Compose makes the multi-container example match the single-container behavior
- the shared `tmpfs` volume provides a runtime area both containers can access
- storing `/run/container` in `tmpfs` keeps the shared state ephemeral

## What This Tutorial Shows

- A multi-process image can be built by layering services across images.
- `.priority` does not control process start order; Nginx and PHP-FPM processes are started concurrently.
- `/run/container` can be used as a shared runtime area between services.
- One image can still be split at runtime with service-selection flags.

## Useful Variants

Run only Nginx:

```bash
docker run --rm -p 8080:80 osixia/nginx-php-fpm-example --exec nginx
```

Run only PHP-FPM:

```bash
docker run --rm osixia/nginx-php-fpm-example --exec php-fpm
```

Debug both services with Bash:

```bash
docker run --rm -it -p 8080:80 osixia/nginx-php-fpm-example --debug
```

## Notes

- The Nginx service still comes from the parent image, including its `startup.sh` behavior and `.env`-driven page customization.
- The PHP-FPM tutorial does not add an `environment/` directory of its own; it relies on what already exists in the parent image.
- If you run this image with a read-only filesystem, `/run/container` must still be writable because the PHP-FPM socket is created there.
- In the compose example, `/run/container` is backed by a shared `tmpfs` volume so Nginx and PHP-FPM can keep using the Unix socket across separate containers.

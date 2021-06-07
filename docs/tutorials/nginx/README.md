# Nginx Tutorial

This tutorial shows how to build a simple Nginx image on top of `osixia/baseimage`.

It demonstrates:

- package installation in the Dockerfile
- service installation with `install.sh`
- runtime customization with `startup.sh`
- a foreground web server process managed through `process.sh`

## Files

### `Dockerfile`

The image:

- starts from `osixia/baseimage`
- installs the `nginx` package during build
- copies the `nginx` service into `/container/services`
- runs `container services install && container services link`
- copies environment files into `/container/environment`

### `environment/.env`

Provides:

```dotenv
CUSTOM_MESSAGE="This page is powered by containers, hot chocolate, and questionable YAML decisions."
```

This value is loaded by the entrypoint before `startup.sh` runs.

### `services/nginx/install.sh`

Executed at build time. It:

- removes the default Debian Nginx welcome page
- creates `/var/www/html/index.html` with the initial content `Hi!`

Because it runs at build time, it does not depend on runtime environment variables.

### `services/nginx/startup.sh`

Executed at container startup before the main process. It:

- checks for `/run/container/nginx-first-start-done`
- appends `${CUSTOM_MESSAGE}` to `/var/www/html/index.html` on first container start only
- writes the marker file so the message is not appended again while the same container filesystem persists

This is a good example of runtime initialization that depends on environment variables.

### `services/nginx/process.sh`

Executed as the managed process. It:

- logs the container IP address
- starts Nginx in the foreground with `daemon off;`
- forwards any extra entrypoint arguments to Nginx

## Build

From this directory:

```bash
docker build -t osixia/nginx-example .
```

## Run

```bash
docker run --rm -p 8080:80 osixia/nginx-example
```

Then open:

```text
http://localhost:8080
```

You should see the generated `Hi!` page with the message from `.env` appended on first startup.

## What This Tutorial Shows

- `install.sh` is used for build-time filesystem preparation.
- `startup.sh` is used for runtime customization from environment variables.
- `process.sh` contains the long-running foreground command.
- `/run/container` is used for runtime state tracking.

## Useful Variants

Run with debug tools and Bash:

```bash
docker run --rm -it -p 8080:80 osixia/nginx-example --debug
```

Run only the process step:

```bash
docker run --rm -p 8080:80 osixia/nginx-example --step process
```

The second command skips `startup.sh`, so the custom message from `.env` is not appended.

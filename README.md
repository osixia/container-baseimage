# osixia/baseimage 🐳✨🌴

[docker hub]: https://hub.docker.com/r/osixia/baseimage
[github]: https://github.com/osixia/container-baseimage

[![Docker Pulls](https://img.shields.io/docker/pulls/osixia/baseimage.svg?style=flat-square)][docker hub]
[![Docker Stars](https://img.shields.io/docker/stars/osixia/baseimage.svg?style=flat-square)][docker hub]
[![GitHub Stars](https://img.shields.io/github/stars/osixia/container-baseimage?label=github%20stars&style=flat-square)][github]
[![Contributors](https://img.shields.io/github/contributors/osixia/container-baseimage?style=flat-square)](https://github.com/osixia/container-baseimage/graphs/contributors)

Debian, Alpine, and Ubuntu container base images designed to build reliable containers quickly.

**This image provides a simple, opinionated solution for building single- or multi-process container images with a minimal number of layers and an optimized build process.**

It accelerates image development and CI/CD pipelines by offering:

 - Advanced build tools to reduce image layers and maximize layer caching efficiency.
 - A lightweight init system as the container entrypoint, providing enhanced process management, debugging capabilities, and runtime flexibility.
 - Built-in support for multi-process containers. Run all processes within a single container or split them across multiple containers as needed.

Includes non-root container images and supports read-only container environments.

![osixia/baseimage logo.](./docs/assets/images/osixia-container-baseimage.jpg)

- [osixia/baseimage 🐳✨🌴](#osixiabaseimage-)
  - [⚡ Quickstart](#-quickstart)
  - [🗂️ Entrypoint Options](#️-entrypoint-options)
  - [📄 Documentation](#-documentation)
  - [🔀 Contributing](#-contributing)
  - [🔓 License](#-license)

## ⚡ Quickstart

Generate image templates in the **osixia-baseimage-example** directory

``` bash
mkdir osixia-baseimage-example
```

``` bash
# Debian
docker run --rm --user $UID --volume $(pwd)/osixia-baseimage-example:/run/container/generator \
osixia/baseimage generate bootstrap
```

``` bash
# Alpine
docker run --rm --user $UID --volume $(pwd)/osixia-baseimage-example:/run/container/generator \
osixia/baseimage:alpine generate bootstrap
```

``` bash
# Ubuntu
docker run --rm --user $UID --volume $(pwd)/osixia-baseimage-example:/run/container/generator \
osixia/baseimage:ubuntu generate bootstrap
```

Note: add `--multi-process` to get a multi-process image sample.

List generated directories and files in **osixia-baseimage-example** directory
``` bash
tree -a osixia-baseimage-example
```

``` bash
osixia-baseimage-example
├── Dockerfile
├── environment
│   ├── .env
│   └── README.md
└── services
    └── service-1
        ├── finish.sh
        ├── install.sh
        ├── .priority
        ├── process.sh
        ├── README.md
        └── startup.sh
```

Take a quick look at the Dockerfile
``` Dockerfile
FROM osixia/baseimage

ARG IMAGE="osixia/baseimage-example:latest"
ENV CONTAINER_IMAGE=${IMAGE}

# Download services required packages or resources
# RUN container packages install --update --clean \
#        [...]
#     && curl -o resources.tar.gz https://[...].tar.gz \
#     && tar -xzf resources.tar.gz \
#     && rm resources.tar.gz

COPY services /container/services

# Install and link services to the entrypoint
RUN container services install \
    && container services link

COPY environment /container/environment
```

Build the image **example/my-image:develop** using files in the **osixia-baseimage-example** directory
``` bash
docker build --tag example/my-image:develop --build-arg IMAGE=example/my-image:develop ./osixia-baseimage-example
```

Note: `--build-arg IMAGE=example/my-image:develop` is used to set the image name inside the container.

Run **example/my-image:develop** image
``` bash
docker run example/my-image:develop
```

``` log
2026-02-26T11:30:59+01:00 INFO    Container image: example/my-image:develop
2026-02-26T11:30:59+01:00 INFO    Loading environment variables from /container/environment/.env ...
2026-02-26T11:30:59+01:00 INFO    Running /container/services/service-1/startup.sh ...
service-1: Doing some container start setup ...
service-1: EXAMPLE_ENV_VAR=Hello :) ...
2026-02-26T11:30:59+01:00 INFO    Running /container/services/service-1/process.sh ...
service-1: Just going to sleep for 10 seconds ...
2026-02-26T11:31:09+01:00 INFO    Running /container/services/service-1/finish.sh ...
service-1: process ended ...
2026-02-26T11:31:09+01:00 INFO    Exiting ...
```

That's it you have a single-process image based on osixia/baseimage.

Detailed documentation regarding the structure, purpose, and usage of the files within the `environment` and `service` directories can be found in their respective `README.md` files.

## 🗂️ Entrypoint Options

``` bash
docker run --rm osixia/baseimage --help
```

```
 / _ \ ___(_)_  _(_) __ _   / / __ )  __ _ ___  ___(_)_ __ ___   __ _  __ _  ___ 
| | | / __| \ \/ / |/ _` | / /|  _ \ / _` / __|/ _ \ | '_ ` _ \ / _` |/ _` |/ _ 
| |_| \__ \ |>  <| | (_| |/ / | |_) | (_| \__ \  __/ | | | | | | (_| | (_| |  __/
 \___/|___/_/_/\_\_|\__,_/_/  |____/ \__,_|___/\___|_|_| |_| |_|\__,_|\__, |\___|
                                                                      |___/      
Container image built with osixia/baseimage (develop) 🐳✨🌴
https://github.com/osixia/container-baseimage

Usage:
  container entrypoint [flags]
  container entrypoint [command]

Aliases:
  entrypoint, ep

Available Commands:
  generate    Generate sample templates

Flags:
  -e, --skip-env-files                      skip loading environment variables from environment files

  -s, --skip-startup                        skip running pre-startup-cmd and startup scripts
  -p, --skip-process                        skip running pre-process-cmd and process scripts
  -f, --skip-finish                         skip running pre-finish-cmd and finish scripts
      --step string                         run only one lifecycle step with its pre-commands and scripts, choices: startup, process, finish

      --pre-startup-cmd stringArray         run command before startup scripts
      --pre-process-cmd stringArray         run command before process scripts
      --pre-finish-cmd stringArray          run command before finish scripts
      --pre-exit-cmd stringArray            run command before container exits

  -x, --exec stringArray                    execute only listed services (default: run services linked to the entrypoint)
  -X, --skip stringArray                    skip listed services
      --skip-all                            skip all services

  -b, --bash                                run Bash alongside other services or command

  -k, --kill-all-on-exit                    kill all remaining container processes on exit (send SIGTERM) (default true)
  -t, --kill-all-on-exit-timeout duration   force kill all processes after timeout (send SIGKILL after SIGTERM timeout) (default 15s)
  -r, --restart                             automatically restart failed processes (single-process: default false, multi-process: default true)
  -a, --keep-alive                          keep container alive after all processes have exited

  -w, --unsafe-fast-write                   disable fsync and related sync operations using eatmydata

  -d, --debug                               set log level to debug, install debug packages, and run Bash
  -i, --install-packages strings            install packages

  -v, --version                             print container image version

  -l, --log-level string                    set log level, choices: error, warning, info, debug, trace (default "info")
      --quiet                               show errors only
  -o, --log-format string                   set log format, choices: console, json (default "console")
  -h, --help                                help for entrypoint

Use "container entrypoint [command] --help" for more information about a command.
```

## 📄 Documentation

See full documentation and complete features list on [osixia/baseimage documentation](https://opensource.osixia.net/projects/container-images/baseimage/).

## 🔀 Contributing

If you find this project useful here's how you can help:

- Send a pull request with new features and bug fixes.
- Help new users with [issues](https://github.com/osixia/container-baseimage/issues) they may encounter.
- Support the development of this image and star [this repo][github] and the image [docker hub repository][docker hub].

This project use a CI/CD tool to build, test and deploy images. See source code and usefull command lines in [build directory](build/).

## 🔓 License

This project is licensed under the terms of the MIT license. See [LICENSE.md](LICENSE.md) file for more information.

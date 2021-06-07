# osixia/baseimage:2.0.0 🐳✨🌴

[docker hub]: https://hub.docker.com/r/osixia/light-baseimage
[github]: https://github.com/osixia/container-baseimage

[![Docker Pulls](https://img.shields.io/docker/pulls/osixia/light-baseimage.svg?style=flat-square)][docker hub]
[![Docker Stars](https://img.shields.io/docker/stars/osixia/light-baseimage.svg?style=flat-square)][docker hub]
[![GitHub Stars](https://img.shields.io/github/stars/osixia/container-baseimage?label=github%20stars&style=flat-square)][github]
[![Contributors](https://img.shields.io/github/contributors/osixia/container-baseimage?style=flat-square)](https://github.com/osixia/container-baseimage/graphs/contributors)

Debian, Alpine and Ubuntu container base images to build reliable images quickly.

**This image provide a simple opinionated solution to build single or multiprocess container images with minimum of layers and an optimized build.**

It helps speeding up image development and CI/CD pipelines by providing:

 - Greats building tools to minimize the image number of layers and make best use of image cache.
 - A nice init process as image entrypoint that add helpfull extensions and options to fastly run and debug containers.
 - Simple way to create multiprocess images.
   Run either all the processes together in a single container or execute them one by one in a Kubernetes pod with multiple containers.

Read-only container filesystem and rootless compatible.

Table of Contents
- [osixia/baseimage:2.0.0 🐳✨🌴](#osixiabaseimage200-)
  - [⚡ Quick Start](#-quick-start)
  - [🗂 Entrypoint Options](#-entrypoint-options)
  - [🍹 First Single Process Image In 2 Minutes](#-first-single-process-image-in-2-minutes)
  - [📄 Documentation](#-documentation)
  - [♥ Contributing](#-contributing)
  - [🔓 License](#-license)
  - [💥 Changelog](#-changelog)

## ⚡ Quick Start

Run the following command to generate a sample Dockerfile and start building an image based on osixia/baseimage:

```
# Debian
docker run --rm osixia/baseimage generate dockerfile --print
```

```
# Alpine
docker run --rm osixia/baseimage:alpine generate dockerfile --print
```

```
# Ubuntu
docker run --rm osixia/baseimage:ubuntu generate dockerfile --print
```

Add `--multiprocess` to get a multiprocess Dockerfile sample. 

Next step: check out a fully functionnal [single process image example](#-first-single-process-image-in-2-minutes).

## 🗂 Entrypoint Options

```
docker run --rm osixia/baseimage --help
```

```
 / _ \ ___(_)_  _(_) __ _   / / __ )  __ _ ___  ___(_)_ __ ___   __ _  __ _  ___ 
| | | / __| \ \/ / |/ _` | / /|  _ \ / _` / __|/ _ \ | '_ ` _ \ / _` |/ _` |/ _ 
| |_| \__ \ |>  <| | (_| |/ / | |_) | (_| \__ \  __/ | | | | | | (_| | (_| |  __/
 \___/|___/_/_/\_\_|\__,_/_/  |____/ \__,_|___/\___|_|_| |_| |_|\__,_|\__, |\___|
                                                                      |___/      
Container image built with osixia/baseimage (2.0.0) 🐳✨🌴
https://github.com/osixia/container-baseimage

Usage:
  container-baseimage entrypoint [flags]
  container-baseimage entrypoint [command]

Aliases:
  entrypoint, ep

Available Commands:
  generate    Generate sample templates
  info        Container image information
  thanks      List container-baseimage contributors

Flags:
  -e, --skip-env-files                      skip getting environment variables values from environment file(s)
                                            
  -s, --skip-startup                        skip running pre-startup-cmd and service(s) startup.sh script(s)
  -p, --skip-process                        skip running pre-process-cmd and service(s) process.sh script(s)
  -f, --skip-finish                         skip running pre-finish-cmd and service(s) finish.sh script(s)
  -c, --run-only-lifecycle-step string      run only one lifecycle step pre-command and script(s) file(s), choices: startup, process, finish
                                            
  -1, --pre-startup-cmd stringArray         run command passed as argument before service(s) startup.sh script(s)
  -3, --pre-process-cmd stringArray         run command passed as argument before service(s) process.sh script(s)
  -5, --pre-finish-cmd stringArray          run command passed as argument before service(s) finish.sh script(s)
  -7, --pre-exit-cmd stringArray            run command passed as argument before container exits
                                            
  -x, --run-only-service string             run only service passed as argument
                                            
  -k, --kill-all-on-exit                    kill all processes on the system upon exiting (send sigterm to all processes) (default true)
  -t, --kill-all-on-exit-timeout duration   kill all processes timeout (send sigkill to all processes after sigterm timeout has been reached) (default 15s)
  -r, --restart-processes                   automatically restart failed services process.sh scripts (multiprocess container images only) (default true)
  -a, --keep-alive                          keep alive container after all processes have exited
                                            
  -w, --unsecure-fast-write                 disable fsync and friends with eatmydata LD_PRELOAD library
                                            
  -d, --debug                               set log level to debug and install debug packages
  -i, --install-packages strings            install packages
                                            
  -v, --version                             print container image version
                                            
  -l, --log-level string                    set log level, choices: none, error, warning, info, debug, trace (default "info")
  -o, --log-format string                   set log format, choices: console, json (default "console")
  -h, --help                                help for entrypoint

Use "container-baseimage entrypoint [command] --help" for more information about a command.
```

## 🍹 First Single Process Image In 2 Minutes
Generate single process image templates in the **osixia-baseimage-example** directory

```
mkdir osixia-baseimage-example
```

```
docker run --rm --user $UID --volume $(pwd)/osixia-baseimage-example:/container/generator/output \
osixia/baseimage generate bootstrap
```

List generated directories and files in **osixia-baseimage-example** directory
```
tree -a osixia-baseimage-example
```

```
osixia-baseimage-example
├── Dockerfile
├── environment
│   └── .env
└── services
    └── service-1
        ├── finish.sh
        ├── install.sh
        ├── .priority
        ├── process.sh
        └── startup.sh
```

Build the image **example/my-image:develop** using files in the **osixia-baseimage-example** directory
```
docker build --tag example/my-image:develop ./osixia-baseimage-example
```

Run **example/my-image:develop** image
```
docker run example/my-image:develop
```

```
2024-01-28T08:21:10Z INFO    Container image: osixia/example-baseimage:latest
2024-01-28T08:21:10Z INFO    Link /container/services/service-1/process.sh to /container/entrypoint/process/service-1/run
2024-01-28T08:21:10Z INFO    Link /container/services/service-1/finish.sh to /container/entrypoint/finish/service-1/run
2024-01-28T08:21:10Z INFO    Link /container/services/service-1/startup.sh to /container/entrypoint/startup/service-1/run
2024-01-28T08:21:10Z INFO    Loading environment variables from /container/environment/.env ...
2024-01-28T08:21:10Z INFO    Increase log level to debug or trace to dump container environment variables
2024-01-28T08:21:10Z INFO    Running script /container/entrypoint/startup/service-1/run ...
service-1: Doing some container first start setup ...
service-1: Doing some others container start setup ...
service-1: EXAMPLE_ENV_VAR=Hello :) ...
2024-01-28T08:21:10Z INFO    Running script /container/entrypoint/process/service-1/run ...
service-1: Just going to sleep for 42 seconds ...

[press ctrl+c to stop process]

2024-01-28T08:21:12Z INFO    Container execution aborted (SIGINT, SIGTERM, SIGQUIT or SIGHUP signal received)
2024-01-28T08:21:12Z INFO    Running script /container/entrypoint/finish/service-1/run ...
service-1: process ended ...
2024-01-28T08:21:12Z INFO    Terminating all processes (timeout: 15s) ...
```

That's it you have a single process image based on osixia/baseimage.

Next steps:
- [Get familiar with generated files]().
- [Customize Dockerfile and service scripts]().
- [Set the container image name to "example/my-image:develop" instead of "osixia/example-baseimage:latest"]().
- [Review image entrypoint options to fastly run and debug containers]().

## 📄 Documentation

⚠ 2.0.0 release is out. Check the [v1 to v2 migration guide](https://opensource.osixia.net/projects/container-images/baseimage/migration-guide-v1-v2/).

See full documentation and complete features list on [osixia/baseimage documentation](https://opensource.osixia.net/projects/container-images/baseimage/).

## ♥ Contributing

If you find this project useful here's how you can help:

- Send a pull request with new features and bug fixes.
- Help new users with [issues](https://github.com/osixia/container-baseimage/issues) they may encounter.
- Support the development of this image and star [this repo][github] and the image [docker hub repository][docker hub].

This project use [dagger](https://github.com/dagger/dagger) as CI/CD tool to build, test and deploy images. See source code and usefull command lines in [ci directory](ci/).

## 🔓 License

This project is licensed under the terms of the MIT license. See [LICENSE.md](LICENSE.md) file for more information.

## 💥 Changelog

Please refer to: [CHANGELOG.md](CHANGELOG.md)

# Service Files
The files outlined below are not mandatory.

## install.sh
This script should exclusively contain instructions for the initial setup of the service.

For improved image construction, all package installations or file downloads should occur within the `Dockerfile`.

By separating time-intensive download operations from the setup, the docker build cache is utilized effectively.
Changes to the `install.sh` file will not necessitate re-downloading dependencies,
as the `Dockerfile` builder will only execute the service installation script.

Note: The `install.sh` script executes during the docker build, thus runtime environment variables cannot be used for setup customization.
Such customizations are handled in the `startup.sh` file.

## startup.sh
This script prepares `process.sh` for execution and tailors the service setup to runtime environment variables.

## process.sh
This script specifies the command to be executed.

For multi-process images, all `process.sh` scripts are launched simultaneously.
The order defined in the service `.priority` file is irrelevant.

## finish.sh
This script is executed once `process.sh` has concluded.

## .priority
The `.priority` file establishes the sequence in which services `install.sh`, `startup.sh` or `finish.sh` scripts are invoked.
The smaller the number, the greater the priority. The default is `500`.

## Optional Service Files

### .optional
This file indicates that the service is optional and is not installed by `container services install` by default.
It can be incorporated later via the `container services require service-1` command.

### download.sh
This script is called during container build to download optional service resources.

Example of a `Dockerfile` using an optional service from a base image:

```
FROM [...]

# Require and download optional services
RUN container services require service-name \
    && container services download

# Install and link services to the entrypoint
RUN container services install \
    && container services link
```

## Service Tags

A service directory may contain a `.tags` sub-directory.
Each file in this directory defines a tag, where the filename is used as the tag name.

Services and their related processes can then be filtered by tag through the services and processes subcommands.

To learn more about tag usage, run `container services status --help`

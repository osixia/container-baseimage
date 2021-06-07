# Service files
All the files described below are optionnal.

## install.sh
This file must only contain directives for the service initial setup.

Files download and apt-get command takes place in the Dockerfile for a better image building experience (or in download.sh file for optional services).

The time consuming download tasks are decoupled from the initial setup to make great use of docker build cache. If an install.sh file is changed the Dockerfile builder will not have to download again dependencies and will just run service(s) install script(s).

Note: The install.sh script is run during the docker build, so run time environment variables can't be used to customize the setup. This is done in the startup.sh file.

## startup.sh
This file is used to make process.sh ready to be run and customize the service setup based on run time environment.

## process.sh
This file define the command to run.

For multiprocess images all the process.sh scripts are started at the same time. Service .priority file do not matter.

## finish.sh
This file is run after process.sh exited.

## .priority
The .priority file defined the order in with services startup.sh or finish.sh scripts are called.
Highter is the number highter is the priority. Default 500.

## .optional
Mark service as optional. The service may be added later by calling services-require command.

## download.sh
This file is called when a optional service is required by services-require command.

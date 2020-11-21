# Changelog

## [1.3.0] - 2020-11-21
### Changed
  - Add loglevel and datetime to log messages
  - Upgrade CFSSL version to 1.5.0

## [1.2.0] - 2020-06-15
### Added
  - Add multiarch support. Thanks to @ndanyluk and @anagno !
  - Allow usage of additional hostnames in self signed certificate #19. Thanks to @Bobonium

### Changed 
  - Use debian buster-slim as baseimage
  - Upgrade python script to python3
  - Upgrade CFSSL version to 1.4.1

### Fixed
  - Fix shellcheck errors and warnings on all scripts

## [1.1.2] - 2019-04-05
### Added
  - jsonssl add support for traefik >= v1.6 acme.json file

### Changed
  - "traefik" JSONSSL_PROFILE be becomes "traefik_up_to_v1_6"
  - "traefik" JSONSSL_PROFILE is now for traefik >= v1.6 acme.json file
  - Upgrade CFSSL version to 1.3.2
  - run: catch copy-service errors
  - KILL_PROCESS_TIMEOUT and KILL_ALL_PROCESSES_TIMEOUT to 30 seconds
  - make ssl-auto-renew cron log with /usr/bin/logger -t cron_ssl_auto_renew
  - syslog-ng config

### Fixed
  - my_init exits with 0 on SIGINT after runit is started
  - better sanitize_shenvname
  - exit status

## [1.1.1] - 2017-10-25
### Changed
  - chmod 444 logrotate config files

### Fixed
  - Fix jsonssl-helper get traefik ca certificate on alpine

## [1.1.0] - 2017-07-19
### Changed
  - Use debian stretch-slim as baseimage

## [1.0.0] - 2017-07-05
### Added
  - Run tool now use 2 environmen variable KILL_PROCESS_TIMEOUT and KILL_ALL_PROCESSES_TIMEOUT

### Changed
  - Default local to en_US.UTF-8

## [0.2.6] - 2016-11-06
### Added
  - Add to the 'run' tool option --dont-touch-etc-hosts Don't add in /etc/hosts a line with the container ip and $HOSTNAME environment variable value.

### Fixed
  - Fix wait-process script

## [0.2.5] - 2016-09-03
### Added
  - Add ssl-helper that allow certificate auto-renew and let choose
    certificate generator (cfssl-helper default, or jsonssl-helper)
  - Add jsonssl-helper that get certificates from a json file
  - Add to the 'run' tool options --run-only, --wait-first-startup, --wait-state, --cmd
   --keepalived becomes --keepalive-force,
   --keepalive now only keep alive container if all startup files and process
     exited without error.

### Changed
  - Upgrade cfssl 1.2.0
  - Change .yaml.startup and .json.startup files to .startup.yaml and .startup.json

### Fixed
  - Fix is_runit_installed check /usr/bin/sv instead of /sbin/runit #6
  - Fix logrotate config

## [0.2.4] - 2016-06-09
### Changed
  - Periodic update of debian baseimage and packages

## [0.2.3] - 2016-05-02
### Changed
  - Periodic update of debian baseimage and packages

## [0.2.2] - 2016-02-20
### Fixed
  - Fix --copy-service error if /container/run/service already exists
  - Fix /container/run/startup.sh file detection if no other startup files exists
  - Fix set_env_hostname_to_etc_hosts() on container restart

## [0.2.1] - 2016-01-25
### Added
  - Add cfssl as available service to generate ssl certs
  - Add tag #PYTHON2BASH and #JSON2BASH to convert env var to bash
  - Add multiple env file importation
  - Add setup only env file
  - Add json env file support
  - Rename my_init to run (delete previous run script)
  - Add run tool option --copy-service that copy /container/service to /container/run/service on startup
  - Add run tool option --loglevel (default : info) with possible values : none, error, warning, info, debug.
  - Add bash log-helper

### Changed
  - Container environment config directory /etc/container_environment moved to /container/environment
  - Container run environment is now saved in /container/run/environment
  - Container run environment bash export /etc/container_environment.sh moved to /container/run/environment.sh
  - Container state is now saved in /container/run/state
  - Container runit process directory /etc/service moved to  /container/run/process
  - Container startup script directory /etc/my_init.d/ moved to /container/run/startup
  - Container final startup script /etc/rc.local moved to /container/run/startup.sh
  - Rename install-multiple-process-stack to add-multiple-process-stack
  - Rename install-service-available to add-service-available

### Removed
  - ssl-helper ssl-helper-openssl and ssl-helper-gnutls
  - Remove run tool option --quiet

## [0.2.0] - 2015-12-16
### Added
  - Makefile with build no cache

### Changed
  - Allow more easy image inheritance

### Fixed
  - Fix cron NUMBER OF HARD LINKS > 1


## [0.1.5] - 2015-11-20
### Fixed
  - Fix bug with host network

## [0.1.4] - 2015-11-19
### Added
  - Add run cmd arguments when it's a single process image

### Changed
  - Remove bash from command when it's a single process image

## [0.1.3] - 2015-11-06
### Added
  - Add hostname env variable to /etc/hosts
    to make the image more friendly with kubernetes again :)

## [0.1.2] - 2015-10-23
### Added
  - Load env.yaml file from /container/environment directory
    to make the image more friendly with kubernetes secrets :)

## [0.1.1] - 2015-08-18
### Added
  - Add python and PyYAML

### Fixed
  - Fix remove-service #1
  - Fix locales
  - Fix my_init

## 0.1.0 - 2015-07-23
Initial release

[1.3.0]: https://github.com/osixia/docker-light-baseimage/compare/v1.2.0...v1.3.0
[1.2.0]: https://github.com/osixia/docker-light-baseimage/compare/v1.1.2...v1.2.0
[1.1.2]: https://github.com/osixia/docker-light-baseimage/compare/v1.1.1...v1.1.2
[1.1.1]: https://github.com/osixia/docker-light-baseimage/compare/v1.1.0...v1.1.1
[1.1.0]: https://github.com/osixia/docker-light-baseimage/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/osixia/docker-light-baseimage/compare/v0.2.2...v1.0.0
[0.2.6]: https://github.com/osixia/docker-light-baseimage/compare/v0.2.5...v0.2.6
[0.2.5]: https://github.com/osixia/docker-light-baseimage/compare/v0.2.4...v0.2.5
[0.2.4]: https://github.com/osixia/docker-light-baseimage/compare/v0.2.3...v0.2.4
[0.2.3]: https://github.com/osixia/docker-light-baseimage/compare/v0.2.2...v0.2.3
[0.2.2]: https://github.com/osixia/docker-light-baseimage/compare/v0.2.1...v0.2.2
[0.2.1]: https://github.com/osixia/docker-light-baseimage/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/osixia/docker-light-baseimage/compare/v0.1.5...v0.2.0
[0.1.5]: https://github.com/osixia/docker-light-baseimage/compare/v0.1.4...v0.1.5
[0.1.4]: https://github.com/osixia/docker-light-baseimage/compare/v0.1.3...v0.1.4
[0.1.3]: https://github.com/osixia/docker-light-baseimage/compare/v0.1.2...v0.1.3
[0.1.2]: https://github.com/osixia/docker-light-baseimage/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/osixia/docker-light-baseimage/compare/v0.1.0...v0.1.1

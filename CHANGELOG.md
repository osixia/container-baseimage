# Changelog

# [0.1.7] - 2021-01-24
### Fixed
  - Update expired default-ca #30 #29. Thanks to @heidemn
  - ARM builds e.g. for keepalived #28. Thanks to @linkvt

# [0.1.6] - 2020-11-21
### Added
  - Add loglevel and datetime to log messages
  - jsonssl add support for traefik >= v1.6 acme.json file
  - Add multiarch support. Thanks to @ndanyluk and @anagno !

### Changed
  - Use alpine:3.12 as baseimage
  - Upgrade python script to python3
  - Upgrade CFSSL version to 1.5.0
  - "traefik" JSONSSL_PROFILE be becomes "traefik_up_to_v1_6"
  - "traefik" JSONSSL_PROFILE is now for traefik >= v1.6 acme.json file
  - run: catch copy-service errors
  - KILL_PROCESS_TIMEOUT and KILL_ALL_PROCESSES_TIMEOUT to 30 seconds
  - make ssl-auto-renew cron log with /usr/bin/logger -t cron_ssl_auto_renew
  - syslog-ng config

### Fixed
  - my_init exits with 0 on SIGINT after runit is started
  - better sanitize_shenvname
  - exit status
  - Fix shellcheck errors and warnings on all scripts

# [0.1.5] - 2017-10-25
### Changed
  - chmod 444 logrotate config files
### Fixed
  - fix jsonssl-helper get traefik ca certificate

## [0.1.4] - 2017-07-19
### Fixed
  - Fix log-helper with piped input

## [0.1.3] - 2017-06-21
### Changed
  - Alpine 3.6

## [0.1.2] - 2017-03-21
### Fixed
  - re-fix ssl-tool package install...

## 0.1.1 - 2017-03-08
### Changed
  - Alpine 3.5
### Fixed
  - Fix ssl-tool package install

## 0.1.0
Initial release

[0.1.7]: https://github.com/osixia/docker-light-baseimage/compare/alpine-v0.1.6...alpine-v0.1.7
[0.1.6]: https://github.com/osixia/docker-light-baseimage/compare/alpine-v0.1.5...alpine-v0.1.6
[0.1.5]: https://github.com/osixia/docker-light-baseimage/compare/alpine-v0.1.4...alpine-v0.1.5
[0.1.4]: https://github.com/osixia/docker-light-baseimage/compare/alpine-v0.1.3...alpine-v0.1.4
[0.1.3]: https://github.com/osixia/docker-light-baseimage/compare/alpine-v0.1.2...alpine-v0.1.3
[0.1.2]: https://github.com/osixia/docker-light-baseimage/compare/alpine-v0.1.1...alpine-v0.1.2

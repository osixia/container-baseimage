# Changelog

# [0.1.6] - Unreleased
### Added
  - jsonssl add support for traefik >= v1.6 acme.json file

### Changed
  - Alpine 3.8
  - "traefik" JSONSSL_PROFILE be becomes "traefik_up_to_v1_6"
  - "traefik" JSONSSL_PROFILE is now for traefik >= v1.6 acme.json file

### Fixed
  - my_init exits with 0 on SIGINT after runit is started
  - better sanitize_shenvname
  - exit status

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

[0.1.6]: https://github.com/osixia/docker-light-baseimage/compare/alpine-v0.1.5...alpine-v0.1.6
[0.1.5]: https://github.com/osixia/docker-light-baseimage/compare/alpine-v0.1.4...alpine-v0.1.5
[0.1.4]: https://github.com/osixia/docker-light-baseimage/compare/alpine-v0.1.3...alpine-v0.1.4
[0.1.3]: https://github.com/osixia/docker-light-baseimage/compare/alpine-v0.1.2...alpine-v0.1.3
[0.1.2]: https://github.com/osixia/docker-light-baseimage/compare/alpine-v0.1.1...alpine-v0.1.2
# Changelog

## 0.1.3
  - Periodic update of debian experimental baseimage and packages

## 0.1.2
  - Fix --copy-service error if /container/run/service already exists
  - Fix /container/run/startup.sh file detection if no other startup files exists
  - Fix set_env_hostname_to_etc_hosts() on container restart

## 0.1.1
  - Add cfssl as available service to generate ssl certs
    Warning: ssl-helper ssl-helper-openssl and ssl-helper-gnutls
             have been removed
  - Add tag #PYTHON2BASH and #JSON2BASH to convert env var to bash
  - Add multiple env file importation
  - Add setup only env file
  - Add json env file support
  - Rename my_init to run (delete previous run script)
  - Add run tool option --copy-service that copy /container/service to /container/run/service on startup
  - Remove run tool option --quiet
  - Add run tool option --loglevel (default : info) with possible values : none, error, warning, info, debug.
  - Container environment config directory /etc/container_environment moved to /container/environment
  - Container run environment is now saved in /container/run/environment
  - Container run environment bash export /etc/container_environment.sh moved to /container/run/environment.sh
  - Container state is now saved in /container/run/state
  - Container runit process directory /etc/service moved to  /container/run/process
  - Container startup script directory /etc/my_init.d/ moved to /container/run/startup
  - Container final startup script /etc/rc.local moved to /container/run/startup.sh
  - Add bash log-helper
  - Rename install-multiple-process-stack to add-multiple-process-stack
  - Rename install-service-available to add-service-available

## 0.1.0
  - Initial release

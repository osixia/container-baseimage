# Changelog

## 1.0.0
  - run tool now use 2 environmen variable KILL_PROCESS_TIMEOUT and KILL_ALL_PROCESSES_TIMEOUT
  - change default local to en_US.UTF-8

## 0.2.6
  - Add to the 'run' tool option --dont-touch-etc-hosts Don't add in /etc/hosts a line with the container ip and $HOSTNAME environment variable value.
  - Fix wait-process script

## 0.2.5
  - Fix is_runit_installed check /usr/bin/sv instead of /sbin/runit #6
  - Upgrade cfssl 1.2.0
  - Add ssl-helper that allow certificate auto-renew and let choose
    certificate generator (cfssl-helper default, or jsonssl-helper)
  - Add jsonssl-helper that get certificates from a json file
  - Add to the 'run' tool options --run-only, --wait-first-startup, --wait-state, --cmd
   --keepalived becomes --keepalive-force,
   --keepalive now only keep alive container if all startup files and process
     exited without error.
  - Change .yaml.startup and .json.startup files to .startup.yaml and .startup.json
  - Fix logrotate config

## 0.2.4
  - Periodic update of debian baseimage and packages

## 0.2.3
  - Periodic update of debian baseimage and packages

## 0.2.2
  - Fix --copy-service error if /container/run/service already exists
  - Fix /container/run/startup.sh file detection if no other startup files exists
  - Fix set_env_hostname_to_etc_hosts() on container restart

## 0.2.1
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

## 0.2.0
  - Allow more easy image inheritance
  - Fix cron NUMBER OF HARD LINKS > 1
  - Makefile with build no cache

## 0.1.5
  - Fix bug with host network

## 0.1.4
  - Add run cmd arguments when it's a single process image
  - Remove bash from command when it's a single process image

## 0.1.3
  - Add hostname env variable to /etc/hosts
    to make the image more friendly with kubernetes again :)

## 0.1.2
  - Load env.yaml file from /container/environment directory
    to make the image more friendly with kubernetes secrets :)

## 0.1.1
  - Fix remove-service #1
  - Add python and PyYAML
  - Fix locales
  - Fix my_init

## 0.1.0
  - Initial release

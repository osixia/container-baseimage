# Changelog

## 0.2.1
  - Add cfssl as available service to generate ssl certs
  - Warning: ssl-helper ssl-helper-openssl and ssl-helper-gnutls are now deprecated
    and will be removed shortly

## 0.2.0
  - Allow more easy image inheritance
  - Fix cron NUMBER OF HARD LINKS > 1
  - Makefile with build no cache

## 0.1.5
  - fix bug with host network

## 0.1.4
  - Add run cmd arguments when it's a single process image
  - Remove bash from command when it's a single process image

## 0.1.3
  - add hostname env variable to /etc/hosts
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

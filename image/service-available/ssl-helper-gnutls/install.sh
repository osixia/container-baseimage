#!/bin/bash -e
./osixia/service-available/ssl-helper/install.sh

LC_ALL=C DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends gnutls-bin

# Fix files permission
chmod 600 /osixia/service-available/ssl-helper-gnutls/certificate-authority/docker_baseimage_gnutls_cakey.pem
chmod 644 /osixia/service-available/ssl-helper-gnutls/certificate-authority/docker_baseimage_gnutls_cacert.pem

# Link certificats et private keys
ln -s /osixia/service-available/ssl-helper-gnutls/certificate-authority/docker_baseimage_gnutls_cacert.pem /etc/ssl/certs/docker_baseimage_gnutls_cacert.pem
ln -s /osixia/service-available/ssl-helper-gnutls/certificate-authority/docker_baseimage_gnutls_cakey.pem /etc/ssl/private/docker_baseimage_gnutls_cakey.pem

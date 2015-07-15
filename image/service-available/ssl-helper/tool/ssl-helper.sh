#!/bin/bash -e

# Usage
# /sbin/ssl-kit crt key [--env-var-prefix=SSL_CRT_] [--ca-crt=/path/to/cert.pem] [--openssl] [--gnutls]

# This script check if files "crt" and "key" exists

# If they do exists :
# - it only fix file permissions

# If not :
# - they will be created usin openssl (default)
#    or gnutls if --gnutls is set in the command line.
# - if --ca-crt is provided it will signed by a CA.
#   and the CA certificat will be linlek to this path

SSL_CRT=$1
SSL_KEY=$2
SSL_CA_CRT=""
SSL_ENV_VAR_PREFIX="SSL_CRT_"
USE_OPENSSL=true

if [ ! -e $SSL_CRT ] || [ ! -e $SSL_KEY ]; then

  echo "Creating files $SSL_CRT and $SSL_KEY"

  # loop script args (skip firts 2)
  for i in ${@:3}
  do

    if [ "${i}" == "---openssl" ]; then
      USE_OPENSSL=true

    elif [ "${i}" == "--gnutls" ]; then
      USE_OPENSSL=false

    elif [[ $i == *--ca-crt* ]]; then
      SSL_CA_CRT=${i#"--ca-crt="}

    elif [[ $i == *--env-var-prefix* ]]; then
      SSL_ENV_VAR_PREFIX=${i#"--env-var-prefix="}
    fi
  done

  # Get ssl vars
  source /osixia/service-available/ssl-helper/tool/get-ssl-env-var.sh
  get_ssl_env_var $SSL_ENV_VAR_PREFIX

  # OPENSSL
  if [ "$USE_OPENSSL" = true ] ; then
    echo "-> Using openssl"

    if [ -z "$SSL_CA_CRT" ]; then
      echo "-> Self signed"

       openssl req -x509 -nodes -days 3650 -newkey rsa:2048 \
               -subj "/C=$SSL_COUNTRY/ST=$SSL_STATE/L=$SSL_LOCATION/O=$SSL_ORGANIZATION/OU=$SSL_ORGANIZATIONAL_UNIT/CN=$SSL_COMMON_NAME$SSL_EMAIL" \
               -keyout $SSL_KEY \
               -out $SSL_CRT

    else
      echo "-> CA signed"

      TMP_FILE="/tmp/docker-cert.csr"

      # CSR
      openssl req -new -nodes \
              -subj "/C=$SSL_COUNTRY/ST=$SSL_STATE/L=$SSL_LOCATION/O=$SSL_ORGANIZATION/OU=$SSL_ORGANIZATIONAL_UNIT/CN=$SSL_COMMON_NAME$SSL_EMAIL" \
              -keyout $SSL_KEY \
              -out $TMP_FILE

      chmod 600 $SSL_KEY

      # Sign request
      openssl x509 -req -in $TMP_FILE -out $SSL_CRT -CA /etc/ssl/certs/docker_baseimage_cacert.pem \
      -CAkey /etc/ssl/private/docker_baseimage_cakey.pem -CAcreateserial -CAserial /etc/ssl/docker_baseimage.srl

      rm $TMP_FILE

      if [ -n "$SSL_CA_CRT" ]; then
        ln -s /etc/ssl/certs/docker_baseimage_cacert.pem $SSL_CA_CRT
      fi
    fi

  # GNUTLS
  else
    echo "-> Using gnutls"

    if [ -z "$SSL_CA_CRT" ]; then
      echo "-> Self signed"
      echo "not supported yet, pull request welcome :)"
      exit 1

    else
      echo "-> CA signed"

      # generate the private key
      certtool --generate-privkey  --sec-param high  --outfile $SSL_KEY

      TMP_FILE="/tmp/docker-cert.cfg"

      source /osixia/service-available/ssl-helper-gnutls/tool/create-gnutls-crt-file-infos.sh
      create_gnutls_crt_file_infos $TMP_FILE

      # create public key
      certtool --generate-certificate --load-privkey $SSL_KEY \
      --load-ca-certificate /etc/ssl/certs/docker_baseimage_gnutls_cacert.pem \
      --load-ca-privkey /etc/ssl/private/docker_baseimage_gnutls_cakey.pem --template $TMP_FILE \
      --outfile $SSL_CRT

      # Delete temp file
      rm $TMP_FILE

      if [ -n "$SSL_CA_CRT" ]; then
        ln -s /etc/ssl/certs/docker_baseimage_gnutls_cacert.pem $SSL_CA_CRT
      fi
    fi
  fi

else
  echo "Files $SSL_CRT and $SSL_KEY already exists"
fi

#fix file permission
chmod 644 $SSL_CRT
chmod 600 $SSL_KEY

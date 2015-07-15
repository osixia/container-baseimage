#!/bin/bash -e

create_gnutls_crt_file_infos() {

	CA_INFOS=$1
	touch $CA_INFOS

	if [ ! -z "$SSL_ORGANIZATION" ]; then
	  echo "organization = \"$SSL_ORGANIZATION\"" >> $CA_INFOS
	fi

	if [ ! -z "$SSL_ORGANIZATIONAL_UNIT" ]; then
	  echo "unit = \"$SSL_ORGANIZATIONAL_UNIT\"" >> $CA_INFOS
	fi

	if [ ! -z "$SSL_LOCATION" ]; then
	  echo "locality = \"$SSL_LOCATION\"" >> $CA_INFOS
	fi

	if [ ! -z "$SSL_STATE" ]; then
	  echo "state = \"$SSL_STATE\"" >> $CA_INFOS
	fi
	if [ ! -z "$SSL_COUNTRY" ]; then
	  echo "country = $SSL_COUNTRY" >> $CA_INFOS
	fi
	if [ ! -z "$SSL_COMMON_NAME" ]; then
	  echo "cn = \"$SSL_COMMON_NAME\"" >> $CA_INFOS
	fi
	if [ ! -z "$SSL_EMAIL" ]; then
	  echo "email = \"$SSL_EMAIL\"" >> $CA_INFOS
	fi

	echo "tls_www_server" >> $CA_INFOS
	echo "encryption_key" >> $CA_INFOS
	echo "signing_key" >> $CA_INFOS
	echo "expiration_days = 3650" >> $CA_INFOS
}
#!/bin/bash -e

# download curl and ca-certificate from apt-get if needed
to_install=()

apk info | grep -q curl || to_install+=("curl")
apk info | grep -q ca-certificates || to_install+=("ca-certificates")

if [ ${#to_install[@]} -ne 0 ]; then
  apk add $to_install
fi

apk add openssl jq

echo "Download cfssl ..."
curl -o /usr/sbin/cfssl -SL https://github.com/osixia/cfssl/releases/download/1.3.3/cfssl_linux-amd64
chmod 700 /usr/sbin/cfssl

echo "Download cfssljson ..."
curl -o /usr/sbin/cfssljson -SL https://github.com/osixia/cfssl/releases/download/1.3.3/cfssljson_linux-amd64
chmod 700 /usr/sbin/cfssljson

echo "Project sources: https://github.com/cloudflare/cfssl"

# remove tools installed to download cfssl
if [ ${#to_install[@]} -ne 0 ]; then
  apk del --purge $to_install
fi

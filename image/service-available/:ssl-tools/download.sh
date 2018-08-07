#!/bin/sh -e

# download curl and ca-certificate from apt-get if needed
to_install=""

apk info | grep -q curl || to_install="curl"
apk info | grep -q ca-certificates || to_install="$to_install ca-certificates"

if [ -n "$to_install" ]; then
  apk add $to_install
fi

apk add openssl jq

echo "Download cfssl ..."
curl -o /usr/sbin/cfssl -SL https://github.com/osixia/cfssl/releases/download/1.3.2/cfssl_linux-amd64
chmod 700 /usr/sbin/cfssl

echo "Download cfssljson ..."
curl -o /usr/sbin/cfssljson -SL https://github.com/osixia/cfssl/releases/download/1.3.2/cfssljson_linux-amd64
chmod 700 /usr/sbin/cfssljson

echo "Project sources: https://github.com/cloudflare/cfssl"

# remove tools installed to download cfssl
if [ -n "$to_install" ]; then
  apk del --purge $to_install
fi

exit 0

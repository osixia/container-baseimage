#!/bin/sh -e

# download curl and ca-certificate from apt-get if needed
to_install=""

apk info curl || to_install="curl"
apk info ca-certificates || to_install="$to_install ca-certificates"

if [ -n "$to_install" ]; then
  apk add $to_install
fi

apk add openssl jq

curl -o /usr/sbin/cfssl -SL https://pkg.cfssl.org/R1.2/cfssl_linux-amd64
chmod 700 /usr/sbin/cfssl

curl -o /usr/sbin/cfssljson -SL https://pkg.cfssl.org/R1.2/cfssljson_linux-amd64
chmod 700 /usr/sbin/cfssljson

# remove tools installed to download cfssl
if [ -n "$to_install" ]; then
  apk del --purge $to_install
fi

exit 0

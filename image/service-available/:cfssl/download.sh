#!/bin/bash -e

# download curl and ca-certificate from apt-get if needed
to_install=""

if [ $(dpkg-query -W -f='${Status}' curl 2>/dev/null | grep -c "ok installed") -eq 0 ]; then
  to_install="curl"
fi

if [ $(dpkg-query -W -f='${Status}' ca-certificates 2>/dev/null | grep -c "ok installed") -eq 0 ]; then
  to_install="$to_install ca-certificates"
fi

if [ -n "$to_install" ]; then
  apk add $to_install
fi

# download libltdl-dev from apt-get
apk add libltdl

curl -o /usr/sbin/cfssl -SL https://github.com/osixia/cfssl/raw/master/bin/alpine/cfssl
chmod 700 /usr/sbin/cfssl

curl -o /usr/sbin/cfssljson -SL https://github.com/osixia/cfssl/raw/master/bin/alpine/cfssljson
chmod 700 /usr/sbin/cfssljson

# remove tools installed to download cfssl
if [ -n "$to_install" ]; then
  apk del --purge $to_install
fi

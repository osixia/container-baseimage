#!/bin/bash -e

# download curl and ca-certificate from apt-get if needed
TO_INSTALL=""

if [ $(dpkg-query -W -f='${Status}' curl 2>/dev/null | grep -c "ok installed") -eq 0 ]; then
  TO_INSTALL="curl"
fi

if [ $(dpkg-query -W -f='${Status}' ca-certificates 2>/dev/null | grep -c "ok installed") -eq 0 ]; then
  TO_INSTALL="$TO_INSTALL ca-certificates"
fi

if [ -n "$TO_INSTALL" ]; then
  LC_ALL=C DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends $TO_INSTALL
fi

# download libltdl-dev from apt-get
LC_ALL=C DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends libltdl-dev

curl -o /usr/sbin/cfssl -SL https://pkg.cfssl.org/R1.1/cfssl_linux-amd64
chmod 700 /usr/sbin/cfssl

curl -o /usr/sbin/cfssljson -SL https://pkg.cfssl.org/R1.1/cfssljson_linux-amd64
chmod 700 /usr/sbin/cfssljson

# remove tools installed to download cfssl
if [ -n "$TO_INSTALL" ]; then
  apt-get remove -y --purge --auto-remove $TO_INSTALL
fi

#!/bin/bash -e

# install curl and ca-certificate id needed
TO_INSTALL=""

if [ $(dpkg-query -W -f='${Status}' curl 2>/dev/null | grep -c "ok installed") -eq 0 ]; then
  TO_INSTALL="curl"
fi

if [ $(dpkg-query -W -f='${Status}' ca-certificates 2>/dev/null | grep -c "ok installed") -eq 0 ]; then
  TO_INSTALL="$TO_INSTALL ca-certificates"
fi

if [ -n "$TO_INSTALL" ]; then
  LC_ALL=C DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends libltdl-dev $TO_INSTALL
fi

curl -o /usr/sbin/cfssl -SL https://github.com/osixia/cfssl/raw/master/bin/cfssl
chmod +x /usr/sbin/cfssl

curl -o /usr/sbin/cfssljson -SL https://github.com/osixia/cfssl/raw/master/bin/cfssljson
chmod +x /usr/sbin/cfssljson

ln -s tool/* /usr/bin

if [ -n "$TO_INSTALL" ]; then
  apt-get remove -y --purge --auto-remove $TO_INSTALL
fi

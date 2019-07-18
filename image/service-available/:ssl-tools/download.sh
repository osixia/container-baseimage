#!/bin/bash -e

case $( dpkg --print-architecture ) in

    "amd64")
    HOST_ARCH="amd64"
    ;;

    "arm64")
    HOST_ARCH="arm64"
    ;;

    "arm" | "armhf")
    HOST_ARCH="arm"
    ;;

    "i386")
    HOST_ARCH="386"
    ;;

    *)
    echo "Unkown architecture. Exiting."
    exit 1
    ;;
esac

echo "The architecture is ${HOST_ARCH}"

# download curl and ca-certificate from apt-get if needed
to_install=()

if [ "$(dpkg-query -W -f='${Status}' curl 2>/dev/null | grep -c "ok installed")" -eq 0 ]; then
    to_install+=("curl")
    echo "Installing curl."
fi

if [ "$(dpkg-query -W -f='${Status}' ca-certificates 2>/dev/null | grep -c "ok installed")" -eq 0 ]; then
    to_install+=("ca-certificates")
    echo "Installing ca-certificates."
fi

if [ ${#to_install[@]} -ne 0 ]; then
    LC_ALL=C DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends "${to_install[@]}"
fi

LC_ALL=C DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends openssl jq

echo "Download cfssl ..."
echo "curl -o /usr/sbin/cfssl -SL https://github.com/osixia/cfssl/releases/download/1.3.4/cfssl_linux-${HOST_ARCH}"
eval curl -o /usr/sbin/cfssl -SL https://github.com/osixia/cfssl/releases/download/1.3.4/cfssl_linux-"${HOST_ARCH}"
chmod 700 /usr/sbin/cfssl

echo "Download cfssljson ..."
echo "curl -o /usr/sbin/cfssljson -SL https://github.com/osixia/cfssl/releases/download/1.3.4/cfssljson_linux-${HOST_ARCH}"
eval curl -o /usr/sbin/cfssljson -SL https://github.com/osixia/cfssl/releases/download/1.3.4/cfssljson_linux-"${HOST_ARCH}"
chmod 700 /usr/sbin/cfssljson

echo "Project sources: https://github.com/cloudflare/cfssl"

# remove tools installed to download cfssl
if [ ${#to_install[@]} -ne 0 ]; then
    apt-get remove -y --purge --auto-remove "${to_install[@]}"
fi

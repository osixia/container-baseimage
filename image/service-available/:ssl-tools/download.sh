#!/bin/bash -e

UARCH=$(uname -m)
echo "Architecture is ${UARCH}"

case "${UARCH}" in
    
    "x86_64")
        HOST_ARCH="amd64"
    ;;
    
    "arm64" | "aarch64")
        HOST_ARCH="arm64"
    ;;
    
    "armv7l" | "armv6l" | "armhf")
        HOST_ARCH="arm"
    ;;
    
    "i386")
        HOST_ARCH="386"
    ;;
    
    *)
        echo "Architecture not supported. Exiting."
        exit 1
    ;;
esac

echo "Going to use ${HOST_ARCH} cfssl binaries"

# download curl and ca-certificate from apt-get if needed
to_install=()

apk info | grep -q curl || to_install+=("curl")
apk info | grep -q ca-certificates || to_install+=("ca-certificates")

if [ ${#to_install[@]} -ne 0 ]; then
    apk add $to_install
fi

apk add openssl jq

echo "Download cfssl ..."
echo "curl -o /usr/sbin/cfssl -SL https://github.com/osixia/cfssl/releases/download/1.5.0/cfssl_linux-${HOST_ARCH}"
curl -o /usr/sbin/cfssl -SL "https://github.com/osixia/cfssl/releases/download/1.5.0/cfssl_linux-${HOST_ARCH}"
chmod 700 /usr/sbin/cfssl

echo "Download cfssljson ..."
echo "curl -o /usr/sbin/cfssljson -SL https://github.com/osixia/cfssl/releases/download/1.5.0/cfssljson_linux-${HOST_ARCH}"
curl -o /usr/sbin/cfssljson -SL "https://github.com/osixia/cfssl/releases/download/1.5.0/cfssljson_linux-${HOST_ARCH}"
chmod 700 /usr/sbin/cfssljson

echo "Project sources: https://github.com/cloudflare/cfssl"

# remove tools installed to download cfssl
if [ ${#to_install[@]} -ne 0 ]; then
    apk del --purge $to_install
fi

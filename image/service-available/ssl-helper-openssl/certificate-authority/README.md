# To create a new CA with openssl
openssl req -new -x509 -nodes -days 3650 -extensions v3_ca -keyout /etc/ssl/private/docker_baseimage_cakey.pem -out /etc/ssl/certs/docker_baseimage_cacert.pem

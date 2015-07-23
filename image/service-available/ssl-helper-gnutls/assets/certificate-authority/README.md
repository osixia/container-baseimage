# To create a new CA with gnutls 
sh -c "certtool --generate-privkey > docker_baseimage_gnutls_cakey.pem"
sudo certtool --generate-self-signed  --load-privkey docker_baseimage_gnutls_cakey.pem --outfile docker_baseimage_gnutls_cacert.pem

Does the certificate belong to an authority? (y/N): -> y
Will the certificate be used to sign other certificates? (y/N): -> y

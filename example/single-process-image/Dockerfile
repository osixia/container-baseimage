# Use osixia/alpine-light-baseimage
# https://github.com/osixia/docker-light-baseimage
FROM osixia/alpine-light-baseimage:0.1.6
MAINTAINER Your Name <your@name.com>

# Download nginx from apk and clean apk files
RUN apk update \
    && apk add \
       nginx \
    && rm -rf /var/cache/apk/*

# Add service directory to /container/service
ADD service /container/service

# Use baseimage install-service script
# https://github.com/osixia/docker-light-baseimage/blob/alpine/image/tool/install-service
RUN /container/tool/install-service

# Add default env directory
ADD environment /container/environment/99-default

# Set /var/lib/nginx/html/ in a data volume
VOLUME /var/lib/nginx/html/

# Expose default http and https ports
EXPOSE 80 443

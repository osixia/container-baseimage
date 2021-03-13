# Use osixia/light-baseimage
# https://github.com/osixia/docker-light-baseimage
FROM osixia/light-baseimage:alpine-0.1.9

# Install multiple process stack, nginx and php7.0-fpm and clean apt-get files
# https://github.com/osixia/docker-light-baseimage/blob/stable/image/tool/add-multiple-process-stack
RUN apk update \
    && /container/tool/add-multiple-process-stack \
    && apk add \
    nginx \
    php7-fpm \
    && rm -rf /var/cache/apk/*

# Add service directory to /container/service
ADD service /container/service

# Use baseimage install-service script
# https://github.com/osixia/docker-light-baseimage/blob/stable/image/tool/install-service
RUN /container/tool/install-service

# Add default env directory
ADD environment /container/environment/99-default

# Set /var/lib/nginx/html/ in a data volume
VOLUME /var/lib/nginx/html/

# Expose default http and https ports
EXPOSE 80 443

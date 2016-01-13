# Use single process image example
FROM example/single-process
MAINTAINER Your Name <your@name.com>

# Install multiple process stack and download php5-fpm
# sources: https://github.com/osixia/docker-light-baseimage/blob/stable/image/tool/install-service-available
#          https://github.com/osixia/docker-light-baseimage/blob/stable/image/service-available/.cfssl/install.sh
RUN echo "deb http://http.debian.net/debian/ jessie main contrib non-free" >> /etc/apt/sources.list \
    && apt-get -y update \
    && /container/tool/install-multiple-process-stack \
    && LC_ALL=C DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
       php5-fpm

# Add service directory to /container/service
# Content :
# /container/service/nginx (from example/single-process)
# /container/service/php5-fpm
ADD service /container/service

# Use baseimage install-service script and clean all
# https://github.com/osixia/docker-light-baseimage/blob/stable/image/tool/install-service
RUN /container/tool/install-service \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

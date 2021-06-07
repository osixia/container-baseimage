FROM osixia/nginx-example

ARG IMAGE="osixia/nginx-php-fpm-example"
ENV CONTAINER_IMAGE=${IMAGE}

# Download service(s) required packages or resources
RUN container packages install --update --clean php-fpm

COPY services /container/services

# Install and link service(s) to the entrypoint
RUN container services install \
    && container services link

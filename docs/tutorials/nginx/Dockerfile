FROM osixia/baseimage

ARG IMAGE="osixia/nginx-example"
ENV CONTAINER_IMAGE=${IMAGE}

# Download services required packages or resources
RUN container packages install --update --clean nginx

COPY services /container/services

# Install and link services to the entrypoint
RUN container services install \
    && container services link

COPY environment /container/environment

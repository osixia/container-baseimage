FROM osixia/baseimage:develop

# Set image name
ARG IMAGE="osixia/baseimage-example:latest"
ENV CONTAINER_IMAGE=${IMAGE}

# Download service(s) required packages or resources
# RUN container packages install --update --clean \
#        [...]
#     && curl -o resources.tar.gz https://[...].tar.gz
#     && tar -xzf resources.tar.gz

COPY services /container/services

# Install and link service(s) to the entrypoint
RUN container services install \
    && container services link

COPY environment /container/environment

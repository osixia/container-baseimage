FROM osixia/baseimage

ARG IMAGE="osixia/single-process-example"
ENV CONTAINER_IMAGE=${IMAGE}

# Download services required packages or resources
# RUN container packages install --update --clean \
#        [...]
#     && curl -o resources.tar.gz https://[...].tar.gz \
#     && tar -xzf resources.tar.gz \
#     && rm resources.tar.gz

COPY services /container/services

# Install and link services to the entrypoint
RUN container services install \
    && container services link

COPY environment /container/environment

FROM pop

# Set container image name
# use to log image version and by templates generator
ARG IMAGE_NAME="osixia/example-multiprocess-baseimage"
ARG IMAGE_TAG="latest"
ENV CONTAINER_IMAGE_NAME=${IMAGE_NAME} \
    CONTAINER_IMAGE_TAG=${IMAGE_TAG}

# Download multiprocess stack, distribution packages
# and others images dependencies
RUN packages-index-update \
    && add-multiprocess-stack
#    && packages-install-clean \
#       [...]

COPY services /container/services

RUN services-install && entrypoint-link-services

COPY environment /container/environment

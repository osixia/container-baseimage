FROM debian:jessie
MAINTAINER Bertrand Gouny <bertrand.gouny@osixia.net>

COPY . /container
RUN /container/build.sh

ENV DEBIAN_FRONTEND teletype

ENV LANG en_US.UTF-8
ENV LANGUAGE en_US:en
ENV LC_ALL en_US.UTF-8

ENTRYPOINT ["/container/tool/run"]

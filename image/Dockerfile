FROM ubuntu:16.04
MAINTAINER Bertrand Gouny <bertrand.gouny@osixia.net>

ADD . /container
RUN /container/build.sh

ENTRYPOINT ["/container/tool/run"]

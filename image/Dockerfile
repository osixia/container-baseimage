FROM ubuntu:trusty
MAINTAINER Bertrand Gouny <bertrand.gouny@osixia.net>

ADD . /container
RUN /container/build.sh

ENTRYPOINT ["/container/tool/run"]

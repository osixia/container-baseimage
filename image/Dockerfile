FROM alpine:3.4
MAINTAINER Bertrand Gouny <bertrand.gouny@osixia.net>

ADD . /container
RUN /container/build.sh

ENTRYPOINT ["/container/tool/run"]

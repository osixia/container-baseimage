FROM ubuntu:trusty
MAINTAINER Bertrand Gouny <bertrand.gouny@osixia.net>

ADD . /container
RUN /container/build.sh

CMD ["/container/tool/run"]

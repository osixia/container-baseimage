FROM debian:jessie
MAINTAINER Bertrand Gouny <bertrand.gouny@osixia.net>

ADD . /osixia
RUN /osixia/build.sh

CMD ["/osixia/tool/run"]

FROM debian:jessie
MAINTAINER Bertrand Gouny <bertrand.gouny@osixia.net>

ENV IS_MULTIPLE_PROCESS false

ADD . /osixia
RUN /osixia/build.sh

CMD ["/sbin/my_init_light"]

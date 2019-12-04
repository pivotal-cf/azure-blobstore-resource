FROM ubuntu:bionic

RUN apt-get update \
  && apt-get install -y --no-install-recommends \
    unzip

ADD assets/ /opt/resource/

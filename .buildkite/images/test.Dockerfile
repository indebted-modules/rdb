FROM golang:1.13.4

RUN apt update && \
  apt install -y postgresql


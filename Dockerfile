FROM golang:1.13.4
RUN go get -u golang.org/x/lint/golint
RUN go get -u golang.org/x/tools/cmd/goimports
RUN go get -u github.com/tj/mmake/cmd/mmake

RUN apt update && \
	apt install -y postgresql

RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.7.0/migrate.linux-amd64.tar.gz | tar xvz
RUN mv migrate.linux-amd64 /usr/local/go/bin/migrate

ENV DOCKERIZE_VERSION v0.6.1
RUN curl -O -L https://github.com/jwilder/dockerize/releases/download/$DOCKERIZE_VERSION/dockerize-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
	&& tar -C /usr/local/bin -xzvf dockerize-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
	&& rm dockerize-linux-amd64-$DOCKERIZE_VERSION.tar.gz

WORKDIR /src

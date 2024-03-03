FROM golang:1.22

LABEL org.opencontainers.image.source=https://github.com/pascalallen/pascalallen.com
LABEL org.opencontainers.image.description="Container image for pascalallen.com"
LABEL org.opencontainers.image.licenses=MIT

WORKDIR /app

ADD . /app

COPY scripts/wait-for-it.sh /usr/bin/wait-for-it.sh

RUN chmod +x /usr/bin/wait-for-it.sh

ENV GOCACHE=/root/.cache/go-build

RUN --mount=type=cache,target="/root/.cache/go-build" CGO_ENABLED=0 GOOS=linux go build -C cmd/pascalallen -o /pascalallen

EXPOSE 9990

ENTRYPOINT /bin/bash /usr/bin/wait-for-it.sh -t 60 $DB_HOST:$DB_PORT -s -- /pascalallen
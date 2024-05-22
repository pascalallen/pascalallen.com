FROM --platform=linux/arm64 golang:1.22

LABEL org.opencontainers.image.source=https://github.com/pascalallen/pascalallen.com
LABEL org.opencontainers.image.description="Container image for pascalallen.com"
LABEL org.opencontainers.image.licenses=MIT

WORKDIR /app

ADD . /app

COPY scripts/wait-for-it.sh /usr/bin/wait-for-it.sh

RUN chmod +x /usr/bin/wait-for-it.sh

ENV GOCACHE=/root/.cache/go-build

RUN --mount=type=cache,target="/root/.cache/go-build" CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -C cmd/pascalallen -o /pascalallen

EXPOSE 9990

CMD /usr/bin/wait-for-it.sh $DB_HOST:$DB_PORT \
    && /usr/bin/wait-for-it.sh $RABBITMQ_HOST:$RABBITMQ_PORT \
    && /pascalallen

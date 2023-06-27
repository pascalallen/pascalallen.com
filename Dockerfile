FROM golang:1.20

LABEL org.opencontainers.image.source=https://github.com/pascalallen/pascalallen.com
LABEL org.opencontainers.image.description="Container image for pascalallen.com"
LABEL org.opencontainers.image.licenses=MIT

RUN mkdir /app

ADD . /app

WORKDIR /app

COPY wait-for-it.sh /usr/bin/wait-for-it.sh

RUN chmod +x /usr/bin/wait-for-it.sh

RUN go build -o /pascalallen

EXPOSE 9990

ENTRYPOINT /bin/bash /usr/bin/wait-for-it.sh -t 60 $DB_HOST:$DB_PORT -s -- /pascalallen
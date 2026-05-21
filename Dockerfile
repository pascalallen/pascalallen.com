FROM --platform=linux/arm64 golang:1.22

LABEL org.opencontainers.image.source=https://github.com/pascalallen/pascalallen.com
LABEL org.opencontainers.image.description="Container image for pascalallen.com"
LABEL org.opencontainers.image.licenses=MIT

WORKDIR /app

ADD . /app

COPY scripts/wait-for-it.sh /usr/bin/wait-for-it.sh

RUN chmod +x /usr/bin/wait-for-it.sh

RUN apt-get update && apt-get install -y --no-install-recommends gnupg2 curl \
    && curl -fsSL https://archive.raspberrypi.com/debian/raspberrypi.gpg.key \
       | gpg --dearmor -o /etc/apt/trusted.gpg.d/raspberrypi.gpg \
    && echo "deb http://archive.raspberrypi.com/debian/ bookworm main" \
       > /etc/apt/sources.list.d/raspi.list \
    && printf 'Package: *\nPin: origin "archive.raspberrypi.com"\nPin-Priority: 1001\n' \
       > /etc/apt/preferences.d/raspberrypi \
    && rm -rf /var/lib/apt/lists/*

RUN apt-get update && apt-get install -y \
    gstreamer1.0-tools \
    gstreamer1.0-plugins-base \
    gstreamer1.0-plugins-good \
    gstreamer1.0-libcamera \
    libcamera-ipa \
    rpicam-apps \
    && rm -rf /var/lib/apt/lists/*

ENV GOCACHE=/root/.cache/go-build

RUN --mount=type=cache,target="/root/.cache/go-build" CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -C cmd/pascalallen -o /pascalallen

EXPOSE 9990

CMD /usr/bin/wait-for-it.sh $DB_HOST:$DB_PORT \
    && /usr/bin/wait-for-it.sh $RABBITMQ_HOST:$RABBITMQ_PORT \
    && /pascalallen

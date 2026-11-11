FROM golang:1.26 AS go-build
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -o /out/pascalallen ./cmd/pascalallen

FROM golang:1.26 AS wasm-build
WORKDIR /src
COPY . .
RUN mkdir -p /out/assets \
    && if [ -f bin/wasm ] && grep -qv '^[[:space:]]*\(#.*\)\?$' bin/wasm; then \
         bash bin/wasm \
         && if [ -d web/static/assets ]; then cp -r web/static/assets/. /out/assets/; fi; \
       fi

FROM node:22 AS web-build
WORKDIR /src
COPY web ./web
COPY --from=wasm-build /out/assets/ ./web/static/assets/
WORKDIR /src/web/app
RUN --mount=type=cache,target=/usr/local/share/.cache/yarn yarn install --frozen-lockfile
RUN yarn build

FROM debian:bookworm-slim AS runtime

ARG REQUIRE_CAMERA=0

LABEL org.opencontainers.image.source=https://github.com/pascalallen/pascalallen.com
LABEL org.opencontainers.image.description="Container image for pascalallen.com"
LABEL org.opencontainers.image.licenses=MIT

RUN apt-get update && apt-get install -y --no-install-recommends \
        ca-certificates \
        curl \
        gnupg2 \
        gstreamer1.0-tools \
        gstreamer1.0-plugins-base \
        gstreamer1.0-plugins-good \
    && if [ "$(dpkg --print-architecture)" = "arm64" ]; then \
         curl -fsSL https://archive.raspberrypi.com/debian/raspberrypi.gpg.key \
           | gpg --dearmor -o /etc/apt/trusted.gpg.d/raspberrypi.gpg \
         && echo "deb http://archive.raspberrypi.com/debian/ bookworm main" \
           > /etc/apt/sources.list.d/raspi.list \
         && printf 'Package: *\nPin: origin "archive.raspberrypi.com"\nPin-Priority: 1001\n' \
           > /etc/apt/preferences.d/raspberrypi \
         && apt-get update \
         && apt-get install -y --no-install-recommends \
              gstreamer1.0-libcamera \
              libcamera-ipa \
              rpicam-apps; \
       fi \
    && if [ "$REQUIRE_CAMERA" = "1" ]; then \
         if [ "$(dpkg --print-architecture)" != "arm64" ] \
            || ! dpkg -s gstreamer1.0-libcamera libcamera-ipa rpicam-apps >/dev/null 2>&1; then \
           echo "REQUIRE_CAMERA=1 but the camera packages are unavailable on $(dpkg --print-architecture); a 64-bit (arm64) userland is required" >&2; \
           exit 1; \
         fi; \
       fi \
    && apt-get purge -y --auto-remove gnupg2 \
    && rm -rf /var/lib/apt/lists/*

RUN useradd --system --create-home --uid 10001 --groups video app

WORKDIR /app
COPY --from=go-build /out/pascalallen /app/pascalallen
COPY web/template /app/web/template
COPY --from=web-build /src/web/static /app/web/static

ENV PORT=9990
EXPOSE 9990

USER app

HEALTHCHECK --interval=30s --timeout=5s --start-period=20s --retries=3 \
    CMD curl -fsS -o /dev/null "http://localhost:${PORT}/" || exit 1

CMD ["/app/pascalallen"]

# rpi-security-cam — Design Spec

**Date:** 2026-05-21  
**Status:** Approved

## Purpose

A self-contained, article-ready demo project that streams live MJPEG video from a Raspberry Pi camera module through a containerized Go web server. Intended to accompany a Medium article on building a homemade security camera with a Pi 5, Pi Camera Module, Ubuntu, and Docker.

---

## Project Layout

```
~/code/rpi-security-cam/
├── Dockerfile
├── compose.yaml
├── go.mod
├── main.go        # HTTP server: routes + listen on :8080
├── hub.go         # GStreamer pub/sub hub
└── index.html     # viewer page, embedded via //go:embed
```

No external Go dependencies. No framework. Pure stdlib + `//go:embed`.

---

## Architecture

### HTTP Server (`main.go`)

- Listens on `:8080`
- Two routes via `net/http`:
  - `GET /` — serves embedded `index.html` with `text/html` content type
  - `GET /stream` — delegates to `hub.ServeStream(w, r)`
- Logs startup message with port

### GStreamer Hub (`hub.go`)

A singleton `Hub` struct that:

- Holds a map of subscriber channels (`map[chan []byte]struct{}`)
- Guards the map with a `sync.Mutex`
- Uses `sync.Once` to start the GStreamer loop on first subscriber
- Runs `gst-launch-1.0` via `exec.Command` with this pipeline:
  ```
  gst-launch-1.0 -q libcamerasrc \
    ! video/x-raw,format=NV12,width=1280,height=720,framerate=15/1 \
    ! videoconvert ! jpegenc quality=75 \
    ! multipartmux boundary=frame ! fdsink fd=1
  ```
- Reads GStreamer stdout in 64 KiB chunks and broadcasts to all subscribers
- On GStreamer exit, sleeps 1 second and restarts (crash recovery)
- Slow/disconnected clients: dropped via non-blocking channel send (no stall)

`ServeStream(w http.ResponseWriter, r *http.Request)`:
- Subscribes a channel, defers unsubscribe
- Sets `Content-Type: multipart/x-mixed-replace; boundary=frame`, `Cache-Control: no-cache`, `X-Accel-Buffering: no`
- Loops: writes chunks from channel to `w`, flushes after each; exits on `r.Context().Done()`

Multiple browser tabs share one GStreamer process — no redundant camera opens.

### Viewer Page (`index.html`)

Single HTML file, no JavaScript:

- Dark (`#111`) background, `100vh` height, flexbox-centered content
- `<h1>` title: "Security Camera"
- `<img src="/stream" alt="camera feed">` — browser-native MJPEG rendering
- Inline `<style>` only, no external CSS

### Dockerfile

```dockerfile
FROM --platform=linux/arm64 golang:1.22

WORKDIR /app
ADD . /app

# Raspberry Pi apt archive (libcamera + GStreamer plugin)
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

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o /rpi-security-cam .

EXPOSE 8080
CMD ["/rpi-security-cam"]
```

### compose.yaml

```yaml
services:
  rpi-security-cam:
    build: .
    container_name: rpi-security-cam
    ports:
      - "8080:8080"
    volumes:
      - /run/udev:/run/udev:ro
    devices:
      - /dev/media0:/dev/media0
      - /dev/media1:/dev/media1
      - /dev/media2:/dev/media2
      - /dev/media3:/dev/media3
      - /dev/video0:/dev/video0
    privileged: true
    restart: unless-stopped
```

---

## Key Constraints

| Constraint | Detail |
|---|---|
| udev mount | `/run/udev:ro` required — libcamera's udev enumerator will find no initialized devices without it |
| NV12 caps | Explicit `video/x-raw,format=NV12` required — libcamera `Camera::configure()` returns EINVAL without it |
| `-q` flag | Suppresses GStreamer state messages that would corrupt the MJPEG stdout stream |
| Exclusive camera access | libcamera allows only one consumer; stop other camera users before running the demo |
| `privileged: true` | Required for `/dev/media*` access inside the container |

---

## Out of Scope

- Authentication / access control
- TLS / HTTPS
- Resolution or framerate configuration via env vars
- Recording to disk
- Motion detection

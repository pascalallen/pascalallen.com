# rpi-security-cam Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a self-contained, containerized Go web server that streams live MJPEG video from a Raspberry Pi camera module to a browser via a single HTML viewer page.

**Architecture:** A singleton `Hub` runs one `gst-launch-1.0` process and fans its MJPEG stdout to N concurrent HTTP clients via buffered Go channels. `main.go` wires two routes (`/` and `/stream`) using stdlib `net/http`. `index.html` is embedded at compile time via `//go:embed`.

**Tech Stack:** Go 1.22 stdlib (`net/http`, `os/exec`, `sync`, `embed`), GStreamer (`gst-launch-1.0` + `gstreamer1.0-libcamera`), Docker + Docker Compose, Raspberry Pi apt archive for `libcamera-ipa`.

---

## File Map

| File | Action | Responsibility |
|---|---|---|
| `go.mod` | Create | Module declaration |
| `hub.go` | Create | GStreamer pub/sub hub + HTTP stream handler |
| `hub_test.go` | Create | Unit tests for subscribe/broadcast/unsubscribe |
| `index.html` | Create | Browser viewer page (dark UI, centered img tag) |
| `main.go` | Create | HTTP server wiring, embed directive, two routes |
| `main_test.go` | Create | HTTP route tests using httptest |
| `Dockerfile` | Create | ARM64 Go build + GStreamer + libcamera-ipa install |
| `compose.yaml` | Create | Device passthrough, udev mount, port binding |
| `.gitignore` | Create | Ignore built binary |

All files live at the project root: `~/code/rpi-security-cam/`.

---

## Task 1: Initialize the project

**Files:**
- Create: `~/code/rpi-security-cam/` (directory)
- Create: `~/code/rpi-security-cam/go.mod`
- Create: `~/code/rpi-security-cam/.gitignore`

- [ ] **Step 1: Create directory and initialize git**

```bash
mkdir ~/code/rpi-security-cam
cd ~/code/rpi-security-cam
git init
```

Expected output:
```
Initialized empty Git repository in /home/pascal/code/rpi-security-cam/.git/
```

- [ ] **Step 2: Initialize Go module**

```bash
cd ~/code/rpi-security-cam
go mod init github.com/pascalallen/rpi-security-cam
```

Expected: creates `go.mod` with:
```
module github.com/pascalallen/rpi-security-cam

go 1.22.0
```

- [ ] **Step 3: Create .gitignore**

Create `~/code/rpi-security-cam/.gitignore`:
```
rpi-security-cam
```

- [ ] **Step 4: Commit**

```bash
cd ~/code/rpi-security-cam
git add go.mod .gitignore
git commit -m "chore: initialize module"
```

---

## Task 2: Implement hub.go (with tests)

**Files:**
- Create: `~/code/rpi-security-cam/hub_test.go`
- Create: `~/code/rpi-security-cam/hub.go`

- [ ] **Step 1: Write failing tests**

Create `~/code/rpi-security-cam/hub_test.go`:
```go
package main

import (
	"testing"
	"time"
)

func TestSubscribeAddsChannel(t *testing.T) {
	h := NewHub()
	ch := h.Subscribe()
	h.mu.Lock()
	_, ok := h.subs[ch]
	h.mu.Unlock()
	if !ok {
		t.Fatal("Subscribe did not register channel in subs map")
	}
}

func TestUnsubscribeRemovesChannel(t *testing.T) {
	h := NewHub()
	ch := h.Subscribe()
	h.Unsubscribe(ch)
	h.mu.Lock()
	_, ok := h.subs[ch]
	h.mu.Unlock()
	if ok {
		t.Fatal("Unsubscribe did not remove channel from subs map")
	}
}

func TestBroadcastDelivers(t *testing.T) {
	h := NewHub()
	ch := make(chan []byte, 1)
	h.mu.Lock()
	h.subs[ch] = struct{}{}
	h.mu.Unlock()

	h.broadcast([]byte("hello"))

	select {
	case msg := <-ch:
		if string(msg) != "hello" {
			t.Errorf("got %q, want %q", string(msg), "hello")
		}
	default:
		t.Fatal("no message received in subscriber channel")
	}
}

func TestBroadcastDropsSlowSubscriber(t *testing.T) {
	h := NewHub()
	// unbuffered channel with no reader simulates a full/slow subscriber
	ch := make(chan []byte)
	h.mu.Lock()
	h.subs[ch] = struct{}{}
	h.mu.Unlock()

	done := make(chan struct{})
	go func() {
		h.broadcast([]byte("drop me"))
		close(done)
	}()

	select {
	case <-done:
		// broadcast returned without blocking — correct
	case <-time.After(time.Second):
		t.Fatal("broadcast blocked on slow subscriber")
	}
}
```

- [ ] **Step 2: Run tests — expect failure**

```bash
cd ~/code/rpi-security-cam
go test ./...
```

Expected: compile error — `NewHub`, `Hub`, etc. undefined.

- [ ] **Step 3: Implement hub.go**

Create `~/code/rpi-security-cam/hub.go`:
```go
package main

import (
	"bytes"
	"log"
	"net/http"
	"os/exec"
	"sync"
	"time"
)

type Hub struct {
	mu   sync.Mutex
	subs map[chan []byte]struct{}
	once sync.Once
}

func NewHub() *Hub {
	return &Hub{subs: make(map[chan []byte]struct{})}
}

func (h *Hub) Subscribe() chan []byte {
	ch := make(chan []byte, 30) // ~2s buffer at 15fps
	h.mu.Lock()
	h.subs[ch] = struct{}{}
	h.mu.Unlock()
	h.once.Do(func() { go h.loop() })
	return ch
}

func (h *Hub) Unsubscribe(ch chan []byte) {
	h.mu.Lock()
	delete(h.subs, ch)
	h.mu.Unlock()
}

func (h *Hub) broadcast(p []byte) {
	chunk := make([]byte, len(p))
	copy(chunk, p)
	h.mu.Lock()
	for ch := range h.subs {
		select {
		case ch <- chunk:
		default: // slow subscriber: drop chunk rather than stall others
		}
	}
	h.mu.Unlock()
}

func (h *Hub) ServeStream(w http.ResponseWriter, r *http.Request) {
	ch := h.Subscribe()
	defer h.Unsubscribe(ch)

	w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary=frame")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case chunk := <-ch:
			if _, err := w.Write(chunk); err != nil {
				return
			}
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}
}

func (h *Hub) loop() {
	for {
		h.runGStreamer()
		time.Sleep(time.Second)
	}
}

func (h *Hub) runGStreamer() {
	cmd := exec.Command("gst-launch-1.0",
		"-q",
		"libcamerasrc",
		"!", "video/x-raw,format=NV12,width=1280,height=720,framerate=15/1",
		"!", "videoconvert",
		"!", "jpegenc", "quality=75",
		"!", "multipartmux", "boundary=frame",
		"!", "fdsink", "fd=1",
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("hub: stdout pipe: %v", err)
		return
	}

	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	if err := cmd.Start(); err != nil {
		log.Printf("hub: gstreamer start: %v", err)
		return
	}
	defer func() {
		cmd.Wait()
		if out := stderrBuf.String(); out != "" {
			log.Printf("hub: gstreamer stderr:\n%s", out)
		}
	}()

	buf := make([]byte, 65536)
	for {
		n, err := stdout.Read(buf)
		if n > 0 {
			h.broadcast(buf[:n])
		}
		if err != nil {
			return
		}
	}
}
```

- [ ] **Step 4: Run tests — expect pass**

```bash
cd ~/code/rpi-security-cam
go test ./... -run TestSubscribe -v
go test ./... -run TestUnsubscribe -v
go test ./... -run TestBroadcast -v
```

Expected output for each:
```
--- PASS: TestSubscribeAddsChannel (0.00s)
--- PASS: TestUnsubscribeRemovesChannel (0.00s)
--- PASS: TestBroadcastDelivers (0.00s)
--- PASS: TestBroadcastDropsSlowSubscriber (0.00s)
PASS
```

- [ ] **Step 5: Commit**

```bash
cd ~/code/rpi-security-cam
git add hub.go hub_test.go
git commit -m "feat: add gstreamer pub/sub hub"
```

---

## Task 3: Create index.html

**Files:**
- Create: `~/code/rpi-security-cam/index.html`

- [ ] **Step 1: Create viewer page**

Create `~/code/rpi-security-cam/index.html`:
```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Security Camera</title>
  <style>
    * { margin: 0; padding: 0; box-sizing: border-box; }
    body {
      background: #111;
      min-height: 100vh;
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      gap: 1rem;
      font-family: sans-serif;
    }
    h1 { color: #eee; font-size: 1.2rem; letter-spacing: 0.1em; text-transform: uppercase; }
    img { max-width: 100%; border-radius: 4px; }
  </style>
</head>
<body>
  <h1>Security Camera</h1>
  <img src="/stream" alt="camera feed">
</body>
</html>
```

- [ ] **Step 2: Commit**

```bash
cd ~/code/rpi-security-cam
git add index.html
git commit -m "feat: add viewer page"
```

---

## Task 4: Implement main.go (with tests)

**Files:**
- Create: `~/code/rpi-security-cam/main_test.go`
- Create: `~/code/rpi-security-cam/main.go`

- [ ] **Step 1: Write failing tests**

Create `~/code/rpi-security-cam/main_test.go`:
```go
package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestIndexReturnsHTML(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	serveIndex(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("got status %d, want %d", rr.Code, http.StatusOK)
	}
	ct := rr.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "text/html") {
		t.Errorf("unexpected Content-Type: %q", ct)
	}
	if !strings.Contains(rr.Body.String(), `<img src="/stream"`) {
		t.Error("response body missing <img src=\"/stream\"> tag")
	}
}

func TestStreamSetsMultipartContentType(t *testing.T) {
	h := NewHub()
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	req := httptest.NewRequest("GET", "/stream", nil).WithContext(ctx)
	rr := httptest.NewRecorder()
	h.ServeStream(rr, req)

	ct := rr.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "multipart/x-mixed-replace") {
		t.Errorf("unexpected Content-Type: %q", ct)
	}
}
```

- [ ] **Step 2: Run tests — expect failure**

```bash
cd ~/code/rpi-security-cam
go test ./...
```

Expected: compile error — `serveIndex` undefined, `indexHTML` undefined.

- [ ] **Step 3: Implement main.go**

Create `~/code/rpi-security-cam/main.go`:
```go
package main

import (
	_ "embed"
	"log"
	"net/http"
)

//go:embed index.html
var indexHTML []byte

var globalHub = NewHub()

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", serveIndex)
	mux.HandleFunc("GET /stream", globalHub.ServeStream)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(indexHTML)
}
```

- [ ] **Step 4: Run all tests — expect pass**

```bash
cd ~/code/rpi-security-cam
go test ./... -v
```

Expected:
```
--- PASS: TestSubscribeAddsChannel (0.00s)
--- PASS: TestUnsubscribeRemovesChannel (0.00s)
--- PASS: TestBroadcastDelivers (0.00s)
--- PASS: TestBroadcastDropsSlowSubscriber (0.00s)
--- PASS: TestIndexReturnsHTML (0.00s)
--- PASS: TestStreamSetsMultipartContentType (0.00s)
PASS
ok  	github.com/pascalallen/rpi-security-cam
```

- [ ] **Step 5: Verify it builds**

```bash
cd ~/code/rpi-security-cam
CGO_ENABLED=0 go build -o /tmp/rpi-security-cam . && echo "build ok"
```

Expected: `build ok`

- [ ] **Step 6: Commit**

```bash
cd ~/code/rpi-security-cam
git add main.go main_test.go
git commit -m "feat: add http server with embed viewer"
```

---

## Task 5: Write Dockerfile

**Files:**
- Create: `~/code/rpi-security-cam/Dockerfile`

- [ ] **Step 1: Create Dockerfile**

Create `~/code/rpi-security-cam/Dockerfile`:
```dockerfile
FROM --platform=linux/arm64 golang:1.22

WORKDIR /app
ADD . /app

# Add Raspberry Pi apt archive — provides gstreamer1.0-libcamera and libcamera-ipa
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

- [ ] **Step 2: Build the image**

```bash
cd ~/code/rpi-security-cam
docker build -t rpi-security-cam .
```

Expected: image builds successfully. The apt steps will take a few minutes on first run. Final line:
```
Successfully tagged rpi-security-cam:latest
```

- [ ] **Step 3: Commit**

```bash
cd ~/code/rpi-security-cam
git add Dockerfile
git commit -m "feat: add dockerfile with gstreamer and libcamera"
```

---

## Task 6: Write compose.yaml

**Files:**
- Create: `~/code/rpi-security-cam/compose.yaml`

- [ ] **Step 1: Create compose.yaml**

Create `~/code/rpi-security-cam/compose.yaml`:
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

> **Note on `/run/udev:ro`:** libcamera's `DeviceEnumeratorUdev` calls `udev_enumerate_add_match_is_initialized()` during camera discovery. Without the host udev database mounted, no devices pass the "initialized" filter and libcamera returns "no cameras found" even though `/dev/media*` are present.

> **Note on `NV12` caps:** The GStreamer pipeline in `hub.go` uses `video/x-raw,format=NV12` between `libcamerasrc` and `videoconvert`. Without explicit format caps, `Camera::configure()` returns EINVAL.

- [ ] **Step 2: Commit**

```bash
cd ~/code/rpi-security-cam
git add compose.yaml
git commit -m "feat: add compose with device passthrough and udev mount"
```

---

## Task 7: On-device smoke test

**Prerequisite:** No other process is currently holding the camera open (libcamera enforces exclusive access). If `pascalallen.com` is running, stop it first: `cd ~/code/pascalallen.com && bin/down prod`.

- [ ] **Step 1: Bring up the stack**

```bash
cd ~/code/rpi-security-cam
docker compose up --build
```

Expected: container starts, then:
```
rpi-security-cam  | 2026/05/21 00:00:00 listening on :8080
```

Shortly after the first browser connection:
```
rpi-security-cam  | WARN: Unsupported V4L2 pixel format RPBP
```
This warning is harmless (Pi ISP raw format unknown to GStreamer).

- [ ] **Step 2: Open the viewer**

Navigate to `http://<pi-ip>:8080` in a browser. The dark-background page should load with a live camera feed in the `<img>` tag. Open a second tab — both should stream simultaneously from the same GStreamer process.

- [ ] **Step 3: Verify crash recovery**

In the compose terminal, note the GStreamer process is running. Kill it from outside:

```bash
pkill gst-launch-1.0
```

Expected: hub logs the exit, sleeps 1 second, restarts GStreamer. Stream resumes in the browser within ~2 seconds.

- [ ] **Step 4: Bring down**

```bash
cd ~/code/rpi-security-cam
docker compose down
```

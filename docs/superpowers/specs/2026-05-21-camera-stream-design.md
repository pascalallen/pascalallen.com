# Camera Stream Design

**Date:** 2026-05-21
**Status:** Approved

## Overview

Add a public `/camera` page to pascalallen.com that streams live video from the onboard Raspberry Pi HQ Camera (IMX477) using MJPEG over HTTP.

## Architecture

A new `Camera()` route group registers `GET /api/v1/camera/stream`. The handler spawns a GStreamer subprocess that captures from the camera via `libcamerasrc`, encodes frames as JPEG, and formats the output as `multipart/x-mixed-replace` on stdout. Go pipes stdout directly to the HTTP response writer. The frontend adds a `/camera` route with a page that renders the stream in a single `<img>` tag.

## Components

### Backend

**`internal/pascalallen/infrastructure/routes/camera.go`** (new)
- Adds `Camera()` method to `Router`, matching the existing pattern in `auth.go` and `default.go`
- Registers `GET /api/v1/camera/stream` → `action.HandleCameraStream()`

**`internal/pascalallen/application/http/action/camera.go`** (new)
- `HandleCameraStream()` returns a `gin.HandlerFunc`
- Sets `Content-Type: multipart/x-mixed-replace; boundary=frame`
- Starts GStreamer subprocess: `gst-launch-1.0 libcamerasrc ! videoconvert ! jpegenc quality=85 ! multipartmux boundary=frame ! fdsink fd=1`
- Copies subprocess stdout to `c.Writer` via `io.Copy`
- Kills subprocess when request context is done (client disconnect)

**`cmd/pascalallen/pascalallen.go`** (modified)
- Add `router.Camera()` call in `configureServer()`

### Frontend

**`web/app/src/domain/constants/Path.ts`** (modified)
- Add `CAMERA = '/camera'`

**`web/app/src/pages/CameraPage.tsx`** (new)
- Functional component with `<Helmet>` title and a single `<img src="/api/v1/camera/stream">`
- Matches the structure of existing page components

**`web/app/src/routes/routes.tsx`** (modified)
- Add public route `{ path: Path.CAMERA, element: <CameraPage /> }`

## Data Flow

```
Browser → GET /camera
  → React Router → CameraPage renders <img src="/api/v1/camera/stream">
  → GET /api/v1/camera/stream
  → HandleCameraStream()
  → gst-launch-1.0 subprocess (libcamerasrc → jpegenc → multipartmux)
  → io.Copy(subprocess stdout → http.ResponseWriter)
  → Browser renders live MJPEG in <img>
```

## Error Handling

- If the GStreamer subprocess fails to start, the handler returns a 500 and logs the error.
- When the client disconnects, the request context is cancelled; the handler detects this and kills the subprocess to avoid orphan processes.
- No retry logic — a page refresh re-initiates the stream.

## Constraints

- Ubuntu 24.04 on Raspberry Pi 5; `rpicam-apps` not available; GStreamer `libcamerasrc` is the capture mechanism.
- Public endpoint — no authentication required.
- Single viewer at a time is expected (home server); no connection limiting needed.
- No new Go dependencies. No new npm packages.

## Files Changed

| Action   | Path |
|----------|------|
| Create   | `internal/pascalallen/infrastructure/routes/camera.go` |
| Create   | `internal/pascalallen/application/http/action/camera.go` |
| Create   | `web/app/src/pages/CameraPage.tsx` |
| Modify   | `cmd/pascalallen/pascalallen.go` |
| Modify   | `web/app/src/domain/constants/Path.ts` |
| Modify   | `web/app/src/routes/routes.tsx` |

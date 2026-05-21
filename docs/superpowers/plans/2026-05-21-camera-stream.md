# Camera Stream Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a public `/camera` page that streams live MJPEG video from the IMX477 HQ Camera via a new `GET /api/v1/camera/stream` endpoint.

**Architecture:** A GStreamer subprocess (`libcamerasrc → videoconvert → jpegenc → multipartmux → fdsink`) writes `multipart/x-mixed-replace` MJPEG data to stdout. The Gin handler pipes that stdout to the HTTP response using `c.Stream`, matching the existing SSE streaming pattern. The frontend renders the stream in a plain `<img>` tag on a new public `/camera` route.

**Tech Stack:** Go 1.26, Gin, GStreamer 1.24 (`libcamerasrc`, `jpegenc`, `multipartmux`, `fdsink`), React 18, TypeScript, React Router, react-helmet-async.

---

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `internal/pascalallen/infrastructure/routes/camera.go` | Register `GET /api/v1/camera/stream` |
| Create | `internal/pascalallen/application/http/action/camera.go` | Spawn GStreamer, pipe to HTTP response |
| Modify | `cmd/pascalallen/pascalallen.go` | Call `router.Camera()` in `configureServer` |
| Modify | `web/app/src/domain/constants/Path.ts` | Add `CAMERA = '/camera'` |
| Create | `web/app/src/pages/CameraPage.tsx` | Page component with `<img>` stream tag |
| Modify | `web/app/src/routes/routes.tsx` | Add public camera route |

---

## Task 1: Camera route registration

**Files:**
- Create: `internal/pascalallen/infrastructure/routes/camera.go`
- Modify: `cmd/pascalallen/pascalallen.go`

- [ ] **Step 1: Create the route file**

Create `internal/pascalallen/infrastructure/routes/camera.go`:

```go
package routes

import "github.com/pascalallen/pascalallen.com/internal/pascalallen/application/http/action"

func (r Router) Camera() {
	v := r.engine.Group(v1)
	{
		camera := v.Group("/camera")
		{
			camera.GET("/stream", action.HandleCameraStream())
		}
	}
}
```

- [ ] **Step 2: Wire the route in `configureServer`**

In `cmd/pascalallen/pascalallen.go`, add `router.Camera()` after `router.Temp(userRepository)`:

```go
func configureServer(container Container) {
	commandBus := container.CommandBus
	userRepository := container.UserRepository

	gin.SetMode(os.Getenv("GIN_MODE"))

	router := routes.NewRouter()

	router.Config()
	router.Fileserver()
	router.Default()
	router.Auth(userRepository, commandBus)
	router.Temp(userRepository)
	router.Camera()
	router.Serve(":9990")
}
```

- [ ] **Step 3: Verify build compiles (action stub needed)**

Since `action.HandleCameraStream()` doesn't exist yet, create a temporary stub to confirm the routing code compiles. Create `internal/pascalallen/application/http/action/camera.go`:

```go
package action

import "github.com/gin-gonic/gin"

func HandleCameraStream() gin.HandlerFunc {
	return func(c *gin.Context) {}
}
```

Run from `/home/pascal/code/pascalallen.com`:

```bash
go build ./...
```

Expected: no output (clean build).

- [ ] **Step 4: Commit**

```bash
git add internal/pascalallen/infrastructure/routes/camera.go \
        internal/pascalallen/application/http/action/camera.go \
        cmd/pascalallen/pascalallen.go
git commit -m "feat: register camera stream route"
```

---

## Task 2: Camera stream action

**Files:**
- Modify: `internal/pascalallen/application/http/action/camera.go`

- [ ] **Step 1: Implement `HandleCameraStream`**

Replace the stub in `internal/pascalallen/application/http/action/camera.go` with the full implementation:

```go
package action

import (
	"io"
	"log"
	"net/http"
	"os/exec"

	"github.com/gin-gonic/gin"
)

func HandleCameraStream() gin.HandlerFunc {
	return func(c *gin.Context) {
		cmd := exec.CommandContext(c.Request.Context(), "gst-launch-1.0",
			"libcamerasrc",
			"!", "videoconvert",
			"!", "jpegenc", "quality=85",
			"!", "multipartmux", "boundary=frame",
			"!", "fdsink", "fd=1",
		)

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Printf("camera stream: stdout pipe: %s", err)
			c.Status(http.StatusInternalServerError)
			return
		}

		if err := cmd.Start(); err != nil {
			log.Printf("camera stream: start: %s", err)
			c.Status(http.StatusInternalServerError)
			return
		}
		defer cmd.Wait()

		c.Header("Content-Type", "multipart/x-mixed-replace; boundary=frame")
		c.Header("Cache-Control", "no-cache")

		buf := make([]byte, 65536)
		c.Stream(func(w io.Writer) bool {
			n, err := stdout.Read(buf)
			if n > 0 {
				if _, writeErr := w.Write(buf[:n]); writeErr != nil {
					return false
				}
			}
			return err == nil
		})
	}
}
```

**Key design notes:**
- `exec.CommandContext` uses `c.Request.Context()` — when the client disconnects, Go cancels the request context, which kills the GStreamer subprocess automatically.
- `defer cmd.Wait()` reaps the subprocess after it exits to avoid zombies.
- `c.Stream` matches the existing pattern in `HandleEventStreamGet` in `temp.go`.
- 64 KB buffer is sized for typical MJPEG frame chunks from the IMX477.

- [ ] **Step 2: Verify build**

```bash
go build ./...
```

Expected: no output (clean build).

- [ ] **Step 3: Integration smoke test**

Start the server (requires Postgres + RabbitMQ — skip if infra not running and proceed to Task 3). If running:

```bash
curl -s --max-time 3 http://localhost:9990/api/v1/camera/stream | head -c 200
```

Expected output starts with:
```
--frame
Content-Type: image/jpeg
```

- [ ] **Step 4: Commit**

```bash
git add internal/pascalallen/application/http/action/camera.go
git commit -m "feat: implement MJPEG camera stream action"
```

---

## Task 3: Frontend — path constant, page, and route

**Files:**
- Modify: `web/app/src/domain/constants/Path.ts`
- Create: `web/app/src/pages/CameraPage.tsx`
- Modify: `web/app/src/routes/routes.tsx`

- [ ] **Step 1: Add `CAMERA` to path constants**

In `web/app/src/domain/constants/Path.ts`, add the `CAMERA` constant:

```typescript
const INDEX = '/';
const LOGIN = '/login';
const FORBIDDEN = '/forbidden';
const TEMP = '/temp';
const CAMERA = '/camera';

export default Object.freeze({
  INDEX,
  LOGIN,
  FORBIDDEN,
  TEMP,
  CAMERA,
});
```

- [ ] **Step 2: Create `CameraPage`**

Create `web/app/src/pages/CameraPage.tsx`:

```tsx
import React, { ReactElement } from 'react';
import { Helmet } from 'react-helmet-async';

const CameraPage = (): ReactElement => {
  return (
    <div className="camera-page">
      <Helmet>
        <title>Pascal Allen - Camera</title>
        <meta name="description" content="Live camera feed" />
      </Helmet>
      <img src="/api/v1/camera/stream" alt="Live camera feed" />
    </div>
  );
};

export default CameraPage;
```

- [ ] **Step 3: Register the route**

In `web/app/src/routes/routes.tsx`, import `CameraPage` and add the public camera route:

```tsx
import React from 'react';
import { RouteObject } from 'react-router-dom';
import Path from '@domain/constants/Path';
import IndexPage from '@pages/IndexPage';
import LoginPage from '@pages/LoginPage';
import TempPage from '@pages/TempPage';
import CameraPage from '@pages/CameraPage';
import RequiresAuthentication from './middleware/RequiresAuthentication';
import RouteElementWrapper from './middleware/RouteElementWrapper';

const routes: RouteObject[] = [
  {
    path: Path.INDEX,
    element: <IndexPage />
  },
  {
    path: Path.LOGIN,
    element: <LoginPage />
  },
  {
    path: Path.TEMP,
    element: (
      <RouteElementWrapper>
        <RequiresAuthentication>
          <TempPage />
        </RequiresAuthentication>
      </RouteElementWrapper>
    )
  },
  {
    path: Path.CAMERA,
    element: <CameraPage />
  }
];

export default routes;
```

- [ ] **Step 4: TypeScript build check**

From `web/app/`:

```bash
yarn dev 2>&1 | tail -20
```

Expected: webpack compilation success, no TypeScript errors.

- [ ] **Step 5: Commit**

```bash
git add web/app/src/domain/constants/Path.ts \
        web/app/src/pages/CameraPage.tsx \
        web/app/src/routes/routes.tsx
git commit -m "feat: add camera page and route"
```

---

## End-to-end verification

After all tasks are complete, verify the full flow:

1. Start the server and navigate to `http://<rpi-ip>:9990/camera`
2. The page should show a live video feed from the camera in the `<img>` element
3. Open browser DevTools → Network tab, confirm the stream request has `Content-Type: multipart/x-mixed-replace; boundary=frame`
4. Navigate away from the page and confirm the GStreamer process exits (check with `pgrep gst-launch`)

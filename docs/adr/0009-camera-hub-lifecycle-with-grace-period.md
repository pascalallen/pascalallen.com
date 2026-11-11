# 0009. The camera hub starts lazily and stops after a grace period

Status: accepted

## Context

The camera process (`gst-launch-1.0`) is expensive on a Pi and the camera is a single-consumer device. Visitors come and go, and a page reload drops and re-adds a subscriber within a second.

## Decision

`infrastructure/camera.Hub` starts the source when the first subscriber arrives. When the last subscriber leaves it waits a grace period (default 5 seconds) before stopping the source, and a subscriber arriving within the grace cancels the stop. A source that exits is restarted with exponential backoff (1 second up to 30 seconds), a start has a 10 second timeout, slow subscribers drop frames instead of blocking the rest, and the stream endpoint answers 503 while the camera is unavailable. `GET /api/v1/camera/status` reports availability. The source and the clock are injected, so the lifecycle is tested without a camera.

## Consequences

- An idle site never holds the camera.
- A reload does not restart the source.
- The host asks the status endpoint before rendering the stream, so an offline camera shows a message instead of a broken image.

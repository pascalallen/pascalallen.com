package action

import (
	"bytes"
	"log"
	"os/exec"
	"sync"
	"time"
)

type cameraHub struct {
	mu   sync.Mutex
	subs map[chan []byte]struct{}
	once sync.Once
}

var hub = &cameraHub{
	subs: make(map[chan []byte]struct{}),
}

func (h *cameraHub) subscribe() chan []byte {
	ch := make(chan []byte, 30) // ~2s buffer at 15fps
	h.mu.Lock()
	h.subs[ch] = struct{}{}
	h.mu.Unlock()
	h.once.Do(func() { go h.loop() })
	return ch
}

func (h *cameraHub) unsubscribe(ch chan []byte) {
	h.mu.Lock()
	delete(h.subs, ch)
	h.mu.Unlock()
}

func (h *cameraHub) broadcast(p []byte) {
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

func (h *cameraHub) loop() {
	for {
		h.runGStreamer()
		time.Sleep(time.Second)
	}
}

func (h *cameraHub) runGStreamer() {
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
		log.Printf("camera hub: stdout pipe: %s", err)
		return
	}

	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	if err := cmd.Start(); err != nil {
		log.Printf("camera hub: start: %s", err)
		return
	}
	defer func() {
		cmd.Wait()
		if out := stderrBuf.String(); out != "" {
			log.Printf("camera hub: gstreamer stderr:\n%s", out)
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

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

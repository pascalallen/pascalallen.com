package action

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func HandleCameraStream() gin.HandlerFunc {
	return func(c *gin.Context) {
		ch := hub.subscribe()
		defer hub.unsubscribe(ch)

		c.Header("Content-Type", "multipart/x-mixed-replace; boundary=frame")
		c.Header("Cache-Control", "no-cache")
		c.Header("X-Accel-Buffering", "no")
		c.Header("Connection", "keep-alive")
		c.Writer.WriteHeader(http.StatusOK)
		c.Writer.Flush()

		ctx := c.Request.Context()
		for {
			select {
			case <-ctx.Done():
				log.Printf("camera stream: client disconnected")
				return
			case chunk := <-ch:
				if _, err := c.Writer.Write(chunk); err != nil {
					log.Printf("camera stream: write error: %s", err)
					return
				}
				c.Writer.Flush()
			}
		}
	}
}

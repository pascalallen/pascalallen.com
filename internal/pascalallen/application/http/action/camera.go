package action

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/pascalallen/pascalallen.com/internal/pascalallen/application/http/responder"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/camera"
)

type cameraStatus struct {
	Available bool `json:"available"`
}

func HandleCameraStatus(stream camera.Stream) gin.HandlerFunc {
	return func(c *gin.Context) {
		responder.OkResponse(c, cameraStatus{Available: stream.Available()})
	}
}

func HandleCameraStream(stream camera.Stream) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		ch, unsubscribe, err := stream.Subscribe(ctx)
		if err != nil {
			if errors.Is(err, camera.ErrUnavailable) {
				responder.ServiceUnavailableResponse(c, err)
			}

			return
		}
		defer unsubscribe()

		c.Header("Content-Type", "multipart/x-mixed-replace; boundary=frame")
		c.Header("Cache-Control", "no-cache")
		c.Header("X-Accel-Buffering", "no")
		c.Header("Connection", "keep-alive")
		c.Writer.WriteHeader(http.StatusOK)
		c.Writer.Flush()

		for {
			select {
			case <-ctx.Done():
				log.Printf("camera stream: client disconnected")
				return
			case chunk, open := <-ch:
				if !open {
					log.Printf("camera stream: source ended")
					return
				}
				if _, err := c.Writer.Write(chunk); err != nil {
					log.Printf("camera stream: write error: %s", err)
					return
				}
				c.Writer.Flush()
			}
		}
	}
}

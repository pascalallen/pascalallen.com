package action

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/application/http/responder"
	"log"
	"mime/multipart"
	"net/textproto"
)

type VideoStreamRequest struct {
	Frames []byte `form:"frames" json:"frames" binding:"required"`
}

func HandleVideoStreamPost(ch chan []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request VideoStreamRequest
		if err := c.ShouldBind(&request); err != nil {
			errorMessage := fmt.Sprintf("request validation error: %s", err.Error())
			responder.BadRequestResponse(c, errors.New(errorMessage))

			return
		}

		ch <- request.Frames

		responder.CreatedResponse(c, &request.Frames)

		return
	}
}

func HandleVideoStreamGet(ch chan []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		mimeWriter := multipart.NewWriter(c.Writer)
		c.Writer.Header().Set("Content-Type", fmt.Sprintf("multipart/x-mixed-replace; boundary=%s", mimeWriter.Boundary()))
		partHeader := make(textproto.MIMEHeader)
		partHeader.Add("Content-Type", "image/jpeg")

		var frame []byte
		for frame = range ch {
			partWriter, err := mimeWriter.CreatePart(partHeader)
			if err != nil {
				log.Printf("failed to create multi-part writer: %s", err)
				return
			}

			if _, err := partWriter.Write(frame); err != nil {
				log.Printf("failed to write image: %s", err)
			}
		}

		return
	}
}

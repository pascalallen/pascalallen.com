package action

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pascalallen/pascalallen.com/http/responder"
	"io"
	"net/http"
)

type EventStreamRequest struct {
	Message string `form:"message" json:"message" binding:"required,max=100"`
}

func HandleTemp() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(
			http.StatusOK,
			responder.JSendSuccessResponse[string]{
				Status: "success",
				Data:   "Ok",
			},
		)

		return
	}
}

func HandleEventStreamPost(ch chan string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request EventStreamRequest
		if err := c.ShouldBind(&request); err != nil {
			errorMessage := fmt.Sprintf("request validation error: %s", err.Error())
			responder.BadRequestResponse(c, errors.New(errorMessage))

			return
		}

		ch <- request.Message

		responder.CreatedResponse(c, &request.Message)

		return
	}
}

func HandleEventStreamGet(ch chan string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Stream(func(w io.Writer) bool {
			if msg, ok := <-ch; ok {
				c.SSEvent("message", msg)
				return true
			}
			return false
		})

		return
	}
}

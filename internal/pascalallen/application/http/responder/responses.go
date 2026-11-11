package responder

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type JSendErrorResponse[T any] struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Code    int    `json:"code,omitempty"`
	Data    T      `json:"data,omitempty"`
}

type JSendFailResponse[T any] struct {
	Status string `json:"status"`
	Data   T      `json:"data"`
}

type JSendSuccessResponse[T any] struct {
	Status string `json:"status"`
	Data   T      `json:"data,omitempty"`
}

func BadRequestResponse(c *gin.Context, error error) {
	c.JSON(
		http.StatusBadRequest,
		JSendFailResponse[string]{
			Status: "fail",
			Data:   error.Error(),
		},
	)

	return
}

func OKResponse[T interface{}](c *gin.Context, i T) {
	c.JSON(
		http.StatusOK,
		JSendSuccessResponse[T]{
			Status: "success",
			Data:   i,
		},
	)

	return
}

func NoContentResponse(c *gin.Context) {
	c.Status(http.StatusNoContent)

	return
}

func PayloadTooLargeResponse(c *gin.Context, message string) {
	c.JSON(
		http.StatusRequestEntityTooLarge,
		JSendFailResponse[string]{
			Status: "fail",
			Data:   message,
		},
	)

	return
}

func BadGatewayResponse(c *gin.Context, message string) {
	c.JSON(
		http.StatusBadGateway,
		JSendErrorResponse[string]{
			Status:  "error",
			Message: message,
			Code:    http.StatusBadGateway,
		},
	)

	return
}

func ServiceUnavailableResponse(c *gin.Context, error error) {
	c.JSON(
		http.StatusServiceUnavailable,
		JSendErrorResponse[string]{
			Status:  "error",
			Message: error.Error(),
		},
	)

	return
}

func NotFoundResponse(c *gin.Context) {
	c.JSON(
		http.StatusNotFound,
		JSendErrorResponse[string]{
			Status:  "error",
			Message: "not found",
		},
	)
}

func MethodNotAllowedResponse(c *gin.Context) {
	c.JSON(
		http.StatusMethodNotAllowed,
		JSendErrorResponse[string]{
			Status:  "error",
			Message: "method not allowed",
		},
	)
}

func OkResponse[T interface{}](c *gin.Context, i T) {
	c.JSON(
		http.StatusOK,
		JSendSuccessResponse[T]{
			Status: "success",
			Data:   i,
		},
	)

	return
}

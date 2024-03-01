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

func InternalServerErrorResponse(c *gin.Context, error error) {
	c.JSON(
		http.StatusInternalServerError,
		JSendErrorResponse[string]{
			Status:  "error",
			Message: error.Error(),
		},
	)

	return
}

func UnprocessableEntityResponse(c *gin.Context, error error) {
	c.JSON(
		http.StatusUnprocessableEntity,
		JSendErrorResponse[string]{
			Status:  "error",
			Message: error.Error(),
		},
	)

	return
}

func UnauthorizedResponse(c *gin.Context, error error) {
	c.JSON(
		http.StatusUnauthorized,
		JSendFailResponse[string]{
			Status: "fail",
			Data:   error.Error(),
		},
	)

	return
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

func CreatedResponse[T interface{}](c *gin.Context, i *T) {
	c.JSON(
		http.StatusCreated,
		JSendSuccessResponse[T]{
			Status: "success",
			Data:   *i,
		},
	)

	return
}

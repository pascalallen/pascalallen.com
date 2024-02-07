package temp

import (
	"github.com/gin-gonic/gin"
	http2 "github.com/pascalallen/pascalallen.com/http"
	"net/http"
)

func HandleTemp() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(
			http.StatusOK,
			http2.JSendSuccessResponse[string]{
				Status: "success",
				Data:   "Ok",
			},
		)

		return
	}
}

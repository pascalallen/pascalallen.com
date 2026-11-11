package action

import (
	"encoding/base64"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
)

func HandleDefault(wasmVersion string) gin.HandlerFunc {
	return func(c *gin.Context) {
		env := map[string]string{
			"APP_BASE_URL": os.Getenv("APP_BASE_URL"),
			"APP_ENV":      os.Getenv("APP_ENV"),
		}
		envBytes, _ := json.Marshal(env)

		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"environment": base64.StdEncoding.EncodeToString(envBytes),
			"wasmVersion": wasmVersion,
		})
	}
}

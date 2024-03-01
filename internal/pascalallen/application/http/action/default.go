package action

import (
	"encoding/base64"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
)

var env = map[string]string{
	"APP_BASE_URL": os.Getenv("APP_BASE_URL"),
	"APP_ENV":      os.Getenv("APP_ENV"),
	"GITHUB_TOKEN": os.Getenv("GITHUB_TOKEN"),
	"SLACK_DM_URL": os.Getenv("SLACK_DM_URL"),
}

func HandleDefault() gin.HandlerFunc {
	return func(c *gin.Context) {
		envBytes, _ := json.Marshal(env)

		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"environment": base64.StdEncoding.EncodeToString(envBytes),
		})

		return
	}
}

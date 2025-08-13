package handlerfunc

import (
	"embed"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/ishaqcherry9/depend/pkg/utils"
)

type CheckHealthReply struct {
	Status   string `json:"status"`
	Hostname string `json:"hostname"`
}

func CheckHealth(c *gin.Context) {
	c.JSON(http.StatusOK, CheckHealthReply{Status: "UP", Hostname: utils.GetHostname()})
}

type PingReply struct{}

func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}

func BrowserRefresh(path string) func(c *gin.Context) {
	return func(c *gin.Context) {
		accept := c.Request.Header.Get("Accept")
		flag := strings.Contains(accept, "text/html")
		if flag {
			content, err := os.ReadFile(path)
			if err != nil {
				c.Writer.WriteHeader(404)
				_, _ = c.Writer.WriteString("Not Found")
				return
			}
			c.Writer.WriteHeader(200)
			c.Writer.Header().Add("Accept", "text/html")
			_, _ = c.Writer.Write(content)
			c.Writer.Flush()
		}
	}
}

func BrowserRefreshFS(fs embed.FS, path string) func(c *gin.Context) {
	return func(c *gin.Context) {
		accept := c.Request.Header.Get("Accept")
		flag := strings.Contains(accept, "text/html")
		if flag {
			content, err := fs.ReadFile(path)
			if err != nil {
				c.Writer.WriteHeader(404)
				_, _ = c.Writer.WriteString("Not Found")
				return
			}
			c.Writer.WriteHeader(200)
			c.Writer.Header().Add("Accept", "text/html")
			_, _ = c.Writer.Write(content)
			c.Writer.Flush()
		}
	}
}

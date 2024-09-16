package copilot

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func agents(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"agents": []interface{}{},
	})
}

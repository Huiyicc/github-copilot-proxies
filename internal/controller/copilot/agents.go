package copilot

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// GetAgents 获取代理列表
func GetAgents(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"agents": []interface{}{},
	})
}

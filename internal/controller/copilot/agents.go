package copilot

import (
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"net/http"
)

// GetAgents 获取代理列表
func GetAgents(c *gin.Context) {
	requestID := uuid.Must(uuid.NewV4()).String()
	c.Header("x-github-request-id", requestID)

	c.JSON(http.StatusOK, gin.H{
		"agents": []interface{}{},
	})
}

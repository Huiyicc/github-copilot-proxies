package copilot

import (
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"net/http"
)

// PostTelemetry 接收并处理来自GitHub Copilot的遥测数据
func PostTelemetry(c *gin.Context) {
	requestID := uuid.Must(uuid.NewV4()).String()
	c.Header("x-github-request-id", requestID)

	c.JSON(http.StatusOK, gin.H{
		"itemsReceived": 0,
		"itemsAccepted": 0,
		"appId":         nil,
		"errors":        []string{},
	})
}

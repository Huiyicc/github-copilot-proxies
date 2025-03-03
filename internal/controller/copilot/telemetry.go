package copilot

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// PostTelemetry 接收并处理来自GitHub Copilot的遥测数据
func PostTelemetry(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"itemsReceived": 0,
		"itemsAccepted": 0,
		"appId":         nil,
		"errors":        []string{},
	})
}

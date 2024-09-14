package copilot

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func postTelemetry(c *gin.Context) {
	var jsonData []interface{}
	if err := c.BindJSON(&jsonData); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"itemsReceived": 0,
			"itemsAccepted": 0,
			"appId":         nil,
			"errors":        []string{err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"itemsReceived": len(jsonData),
		"itemsAccepted": len(jsonData),
		"appId":         nil,
		"errors":        []string{},
	})
}

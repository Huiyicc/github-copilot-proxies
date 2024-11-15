package copilot

import (
	"github.com/gin-gonic/gin"
)

func HandleEmbeddings(c *gin.Context) {
	var req EmbeddingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 设置默认值
	if req.Dimensions == 0 {
		req.Dimensions = dimensionSize
	}

	resp, err := getEmbeddings(req.Input, req.Dimensions)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, resp)
}

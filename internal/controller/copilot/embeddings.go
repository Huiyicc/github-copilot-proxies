package copilot

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// EmbeddingsAPIRequest 表示嵌入API的请求结构
type EmbeddingsAPIRequest struct {
	Input      []string `json:"input" binding:"required"`
	Model      string   `json:"model,omitempty"`
	Dimensions int      `json:"dimensions,omitempty"`
}

// HandleEmbeddings 处理嵌入请求的HTTP处理器
func HandleEmbeddings(c *gin.Context) {
	var req EmbeddingsAPIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 创建嵌入客户端
	client, err := NewEmbeddingClient(dimensionSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 如果请求中指定了模型，则使用请求中的模型
	if req.Model != "" {
		client.SetModel(req.Model)
	}

	// 获取嵌入，使用请求上下文以支持取消操作
	resp, err := client.GetEmbeddings(c.Request.Context(), req.Input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

package copilot

import (
	"github.com/gofrs/uuid"
	"net/http"
	"os"
	"strconv"

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
	requestID := uuid.Must(uuid.NewV4()).String()
	c.Header("x-github-request-id", requestID)

	var req EmbeddingsAPIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 从环境变量获取维度大小，默认为1536
	dimensionSize := 1536
	if dimSizeStr := os.Getenv("EMBEDDING_DIMENSION_SIZE"); dimSizeStr != "" {
		if dimSize, err := strconv.Atoi(dimSizeStr); err == nil {
			dimensionSize = dimSize
		}
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

// getEmbeddingModels 获取可用的嵌入模型列表
func getEmbeddingModels(c *gin.Context) {
	modelName := os.Getenv("EMBEDDING_API_MODEL_NAME")
	if modelName == "" {
		modelName = "text-embedding-3-small"
	}

	requestID := uuid.Must(uuid.NewV4()).String()
	c.Header("x-github-request-id", requestID)
	c.JSON(http.StatusOK, gin.H{
		"data": []gin.H{
			{"id": modelName, "object": "model", "owned_by": "openai", "permission": []string{}},
		},
		"object": "list",
	})
}

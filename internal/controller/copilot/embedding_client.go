package copilot

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

// 常量定义
const (
	defaultEmbeddingModel = "text-embedding-v3"
	defaultTimeout        = 30 * time.Second
	contentTypeJSON       = "application/json"
)

// EmbeddingRequest 表示向嵌入API发送的请求
type EmbeddingRequest struct {
	Model      string   `json:"model"`
	Input      []string `json:"input"`
	Dimensions int      `json:"dimensions"`
}

// EmbeddingResponse 表示从嵌入API接收的响应
type EmbeddingResponse struct {
	Data   []EmbeddingData `json:"data"`
	Model  string          `json:"model"`
	Object string          `json:"object"`
	Usage  Usage           `json:"usage"`
}

// EmbeddingData 表示单个嵌入数据
type EmbeddingData struct {
	Embedding []float32 `json:"embedding"`
	Index     int       `json:"index"`
	Object    string    `json:"object"`
}

// Usage 表示API使用情况
type Usage struct {
	PromptTokens int `json:"prompt_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// 移除未使用的类型
// Parameters 和 EmbeddingsRequest, EmbeddingsResponse 已被移除

// EmbeddingClient 封装了与嵌入API交互的功能
type EmbeddingClient struct {
	apiURL      string
	apiKey      string
	model       string
	dimensions  int
	httpClient  *http.Client
	clientMutex sync.RWMutex
}

// NewEmbeddingClient 创建一个新的嵌入客户端
func NewEmbeddingClient(dimensions int) (*EmbeddingClient, error) {
	apiURL := os.Getenv("EMBEDDING_API_BASE")
	apiKey := os.Getenv("EMBEDDING_API_KEY")

	if apiURL == "" || apiKey == "" {
		return nil, fmt.Errorf("EMBEDDING_API_BASE or EMBEDDING_API_KEY environment variable not set")
	}

	// 解析超时时间，如果未设置或解析失败则使用默认值
	timeout := defaultTimeout
	if timeoutStr := os.Getenv("HTTP_CLIENT_TIMEOUT"); timeoutStr != "" {
		if parsedTimeout, err := time.ParseDuration(timeoutStr + "s"); err == nil {
			timeout = parsedTimeout
		}
	}

	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	return &EmbeddingClient{
		apiURL:     apiURL,
		apiKey:     apiKey,
		model:      defaultEmbeddingModel,
		dimensions: dimensions,
		httpClient: client,
	}, nil
}

// SetModel 设置嵌入模型
func (c *EmbeddingClient) SetModel(model string) {
	c.clientMutex.Lock()
	defer c.clientMutex.Unlock()
	c.model = model
}

// GetEmbedding 获取单个文本的嵌入
func (c *EmbeddingClient) GetEmbedding(ctx context.Context, text string) ([]float32, error) {
	resp, err := c.GetEmbeddings(ctx, []string{text})
	if err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}

	return resp.Data[0].Embedding, nil
}

// GetEmbeddings 批量获取多个文本的嵌入
func (c *EmbeddingClient) GetEmbeddings(ctx context.Context, texts []string) (*EmbeddingResponse, error) {
	c.clientMutex.RLock()
	model := c.model
	dimensions := c.dimensions
	c.clientMutex.RUnlock()

	reqBody := EmbeddingRequest{
		Model:      model,
		Input:      texts,
		Dimensions: dimensions,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", contentTypeJSON)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var embeddingResp EmbeddingResponse
	if err := json.Unmarshal(body, &embeddingResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	return &embeddingResp, nil
}

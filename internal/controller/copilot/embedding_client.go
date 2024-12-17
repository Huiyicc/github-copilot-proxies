package copilot

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// 接口文档: https://help.aliyun.com/zh/model-studio/developer-reference/text-embedding-synchronous-api?spm=a2c4g.11186623.0.0.40ed7ee8SeSJx7

const (
	embeddingURL = "https://dashscope.aliyuncs.com/api/v1/services/embeddings/text-embedding/text-embedding"
)

type EmbeddingRequest struct {
	Model      string      `json:"model"`
	Input      InputParams `json:"input"`
	Parameters Parameters  `json:"parameters"`
}

type InputParams struct {
	Texts []string `json:"texts"`
}

type Parameters struct {
	Dimension  int    `json:"dimension"`
	TextType   string `json:"text_type,omitempty"`
	OutputType string `json:"output_type,omitempty"`
}

type EmbeddingResponse struct {
	Output struct {
		Embeddings []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"embeddings"`
	} `json:"output"`
	Usage struct {
		InputTokens int `json:"input_tokens"`
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
}

type EmbeddingsRequest struct {
	Input      []string `json:"input"`
	Model      string   `json:"model"`
	Dimensions int      `json:"dimensions"`
	TextType   string   `json:"text_type,omitempty"`
	OutputType string   `json:"output_type,omitempty"`
}

type EmbeddingData struct {
	Embedding []float32 `json:"embedding"`
	Index     int       `json:"index"`
}

type Usage struct {
	InputTokens int `json:"input_tokens"`
	TotalTokens int `json:"total_tokens"`
}

type EmbeddingsResponse struct {
	Data  []EmbeddingData `json:"data"`
	Usage Usage           `json:"usage"`
}

func getEmbedding(text string) ([]float32, error) {
	apiKey := os.Getenv("DASHSCOPE_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("DASHSCOPE_API_KEY environment variable not set")
	}

	reqBody := EmbeddingRequest{
		Model: "text-embedding-v2",
		Input: InputParams{
			Texts: []string{text},
		},
		Parameters: Parameters{
			Dimension:  dimensionSize,
			TextType:   "document",
			OutputType: "dense",
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", embeddingURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	httpClientTimeout, _ := time.ParseDuration(os.Getenv("HTTP_CLIENT_TIMEOUT") + "s")
	client := &http.Client{
		Timeout: httpClientTimeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	resp, err := client.Do(req)
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

	if len(embeddingResp.Output.Embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}

	return embeddingResp.Output.Embeddings[0].Embedding, nil
}

func getEmbeddings(texts []string, dimension int) (*EmbeddingsResponse, error) {
	apiKey := os.Getenv("DASHSCOPE_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("DASHSCOPE_API_KEY environment variable not set")
	}

	reqBody := EmbeddingRequest{
		Model: "text-embedding-v3",
		Input: InputParams{
			Texts: texts,
		},
		Parameters: Parameters{
			Dimension:  dimension,
			TextType:   "document",
			OutputType: "dense",
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", embeddingURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	httpClientTimeout, _ := time.ParseDuration(os.Getenv("HTTP_CLIENT_TIMEOUT") + "s")
	client := &http.Client{
		Timeout: httpClientTimeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	resp, err := client.Do(req)
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

	result := &EmbeddingsResponse{
		Data: make([]EmbeddingData, len(embeddingResp.Output.Embeddings)),
		Usage: Usage{
			InputTokens: embeddingResp.Usage.InputTokens,
			TotalTokens: embeddingResp.Usage.TotalTokens,
		},
	}

	for i, emb := range embeddingResp.Output.Embeddings {
		result.Data[i] = EmbeddingData{
			Embedding: emb.Embedding,
			Index:     i,
		}
	}

	return result, nil
}

package copilot

import (
	_ "embed"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"time"
)

type Pong struct {
	Now    int    `json:"now"`
	Status string `json:"status"`
	Ns1    string `json:"ns1"`
}

// GetPing 模拟ping接口
func GetPing(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, Pong{
		Now:    time.Now().Second(),
		Status: "ok",
		Ns1:    "200 OK",
	})
}

// GetModels 获取模型列表
func GetModels(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"data": []gin.H{
			{
				"capabilities": gin.H{
					"family":    "gpt-3.5-turbo",
					"limits":    gin.H{"max_prompt_tokens": 12288, "max_context_window_tokens": 16384, "max_output_tokens": 4096},
					"object":    "model_capabilities",
					"supports":  gin.H{"tool_calls": true},
					"tokenizer": "cl100k_base",
					"type":      "chat",
				},
				"id":                   "gpt-3.5-turbo",
				"name":                 "GPT 3.5 Turbo",
				"object":               "model",
				"version":              "gpt-3.5-turbo-0613",
				"model_picker_enabled": false,
			},
			{
				"capabilities": gin.H{
					"family":    "gpt-3.5-turbo",
					"limits":    gin.H{"max_prompt_tokens": 12288, "max_context_window_tokens": 16384, "max_output_tokens": 4096},
					"object":    "model_capabilities",
					"supports":  gin.H{"tool_calls": true},
					"tokenizer": "cl100k_base",
					"type":      "chat",
				},
				"id":                   "gpt-3.5-turbo-0613",
				"name":                 "GPT 3.5 Turbo",
				"object":               "model",
				"version":              "gpt-3.5-turbo-0613",
				"model_picker_enabled": false,
			},
			{
				"capabilities": gin.H{
					"family":    "gpt-4",
					"limits":    gin.H{"max_prompt_tokens": 32768, "max_context_window_tokens": 32768, "max_output_tokens": 4096},
					"object":    "model_capabilities",
					"supports":  gin.H{"tool_calls": true},
					"tokenizer": "cl100k_base",
					"type":      "chat",
				},
				"id":                   "gpt-4",
				"name":                 "GPT 4",
				"object":               "model",
				"version":              "gpt-4-0613",
				"model_picker_enabled": false,
			},
			{
				"capabilities": gin.H{
					"family":    "gpt-4",
					"limits":    gin.H{"max_prompt_tokens": 32768, "max_context_window_tokens": 32768, "max_output_tokens": 4096},
					"object":    "model_capabilities",
					"supports":  gin.H{"tool_calls": true},
					"tokenizer": "cl100k_base",
					"type":      "chat",
				},
				"id":                   "gpt-4-0613",
				"name":                 "GPT 4",
				"object":               "model",
				"version":              "gpt-4-0613",
				"model_picker_enabled": false,
			},
			{
				"capabilities": gin.H{
					"family":    "gpt-4o-mini",
					"limits":    gin.H{"max_prompt_tokens": 12288, "max_context_window_tokens": 128000, "max_output_tokens": 4096},
					"object":    "model_capabilities",
					"supports":  gin.H{"tool_calls": true, "parallel_tool_calls": true},
					"tokenizer": "o200k_base",
					"type":      "chat",
				},
				"id":                   "gpt-4o-mini",
				"name":                 "GPT 4o Mini",
				"object":               "model",
				"version":              "gpt-4o-mini-2024-07-18",
				"model_picker_enabled": false,
			},
			{
				"capabilities": gin.H{
					"family":    "gpt-4o-mini",
					"limits":    gin.H{"max_prompt_tokens": 12288, "max_context_window_tokens": 128000, "max_output_tokens": 4096},
					"object":    "model_capabilities",
					"supports":  gin.H{"tool_calls": true, "parallel_tool_calls": true},
					"tokenizer": "o200k_base",
					"type":      "chat",
				},
				"id":                   "gpt-4o-mini-2024-07-18",
				"name":                 "GPT 4o Mini",
				"object":               "model",
				"version":              "gpt-4o-mini-2024-07-18",
				"model_picker_enabled": false,
			},
			{
				"capabilities": gin.H{
					"family":    "gpt-4-turbo",
					"limits":    gin.H{"max_prompt_tokens": 64000, "max_context_window_tokens": 128000, "max_output_tokens": 4096},
					"object":    "model_capabilities",
					"supports":  gin.H{"parallel_tool_calls": true, "tool_calls": true},
					"tokenizer": "cl100k_base",
					"type":      "chat",
				},
				"id":                   "gpt-4-0125-preview",
				"name":                 "GPT 4 Turbo",
				"object":               "model",
				"version":              "gpt-4-0125-preview",
				"model_picker_enabled": false,
			},
			{
				"capabilities": gin.H{
					"family":    "gpt-4o",
					"limits":    gin.H{"max_prompt_tokens": 64000, "max_context_window_tokens": 128000, "max_output_tokens": 4096},
					"object":    "model_capabilities",
					"supports":  gin.H{"parallel_tool_calls": true, "tool_calls": true},
					"tokenizer": "o200k_base",
					"type":      "chat",
				},
				"id":                   "gpt-4o",
				"name":                 "GPT 4o",
				"object":               "model",
				"version":              "gpt-4o-2024-05-13",
				"model_picker_enabled": true,
			},
			{
				"capabilities": gin.H{
					"family":    "gpt-4o",
					"limits":    gin.H{"max_prompt_tokens": 64000, "max_context_window_tokens": 128000, "max_output_tokens": 4096},
					"object":    "model_capabilities",
					"supports":  gin.H{"parallel_tool_calls": true, "tool_calls": true},
					"tokenizer": "o200k_base",
					"type":      "chat",
				},
				"id":                   "gpt-4o-2024-05-13",
				"name":                 "GPT 4o",
				"object":               "model",
				"version":              "gpt-4o-2024-05-13",
				"model_picker_enabled": false,
			},
			{
				"capabilities": gin.H{
					"family":    "gpt-4o",
					"limits":    gin.H{"max_prompt_tokens": 64000, "max_context_window_tokens": 128000, "max_output_tokens": 4096},
					"object":    "model_capabilities",
					"supports":  gin.H{"parallel_tool_calls": true, "tool_calls": true},
					"tokenizer": "o200k_base",
					"type":      "chat",
				},
				"id":                   "gpt-4-o-preview",
				"name":                 "GPT 4o",
				"object":               "model",
				"version":              "gpt-4o-2024-05-13",
				"model_picker_enabled": false,
			},
			{
				"capabilities": gin.H{
					"family":    "gpt-4o",
					"limits":    gin.H{"max_prompt_tokens": 64000, "max_context_window_tokens": 128000, "max_output_tokens": 4096},
					"object":    "model_capabilities",
					"supports":  gin.H{"parallel_tool_calls": true, "tool_calls": true},
					"tokenizer": "o200k_base",
					"type":      "chat",
				},
				"id":                   "gpt-4o-2024-08-06",
				"name":                 "GPT 4o",
				"object":               "model",
				"version":              "gpt-4o-2024-08-06",
				"model_picker_enabled": false,
			},
			{
				"capabilities": gin.H{
					"family":    "o1",
					"limits":    gin.H{"max_prompt_tokens": 64000, "max_context_window_tokens": 128000},
					"object":    "model_capabilities",
					"supports":  gin.H{},
					"tokenizer": "o200k_base",
					"type":      "chat",
				},
				"id":                   "o1-preview",
				"name":                 "o1-preview (Preview)",
				"object":               "model",
				"version":              "gpt-4o-2024-08-06",
				"model_picker_enabled": true,
			},
			{
				"capabilities": gin.H{
					"family":    "o1",
					"limits":    gin.H{"max_prompt_tokens": 64000, "max_context_window_tokens": 128000},
					"object":    "model_capabilities",
					"supports":  gin.H{},
					"tokenizer": "o200k_base",
					"type":      "chat",
				},
				"id":                   "o1-preview-2024-09-12",
				"name":                 "o1-preview (Preview)",
				"object":               "model",
				"version":              "o1-preview-2024-09-12",
				"model_picker_enabled": false,
			},
			{
				"capabilities": gin.H{
					"family":    "o1-mini",
					"limits":    gin.H{"max_prompt_tokens": 20000, "max_context_window_tokens": 128000},
					"object":    "model_capabilities",
					"supports":  gin.H{},
					"tokenizer": "o200k_base",
					"type":      "chat",
				},
				"id":                   "o1-mini",
				"name":                 "o1-mini (Preview)",
				"object":               "model",
				"version":              "o1-preview-2024-09-12",
				"model_picker_enabled": true,
			},
			{
				"capabilities": gin.H{
					"family":    "o1-mini",
					"limits":    gin.H{"max_prompt_tokens": 20000, "max_context_window_tokens": 128000},
					"object":    "model_capabilities",
					"supports":  gin.H{},
					"tokenizer": "o200k_base",
					"type":      "chat",
				},
				"id":                   "o1-mini-2024-09-12",
				"name":                 "o1-mini (Preview)",
				"object":               "model",
				"version":              "o1-mini-2024-09-12",
				"model_picker_enabled": false,
			},
			{
				"capabilities": gin.H{
					"family":    "claude-3.5-sonnet",
					"limits":    gin.H{"max_prompt_tokens": 195000, "max_context_window_tokens": 200000, "max_output_tokens": 4096},
					"object":    "model_capabilities",
					"supports":  gin.H{},
					"tokenizer": "o200k_base",
					"type":      "chat",
				},
				"id":                   "claude-3.5-sonnet",
				"name":                 "Claude 3.5 Sonnet (Preview)",
				"object":               "model",
				"version":              "claude-3.5-sonnet",
				"model_picker_enabled": true,
				"policy": gin.H{
					"state": "enabled",
					"terms": "Enable access to the latest Claude 3.5 Sonnet model from Anthropic. [Learn more about how GitHub Copilot serves Claude 3.5 Sonnet](https://docs.github.com/copilot/using-github-copilot/using-claude-sonnet-in-github-copilot).",
				},
			},
			{
				"capabilities": gin.H{
					"family":    "text-embedding-ada-002",
					"limits":    gin.H{"max_inputs": 256},
					"object":    "model_capabilities",
					"supports":  gin.H{},
					"tokenizer": "cl100k_base",
					"type":      "embeddings",
				},
				"id":                   "text-embedding-ada-002",
				"name":                 "Embedding V2 Ada",
				"object":               "model",
				"version":              "text-embedding-ada-002",
				"model_picker_enabled": false,
			},
			{
				"capabilities": gin.H{
					"family":    "text-embedding-3-small",
					"limits":    gin.H{"max_inputs": 256},
					"object":    "model_capabilities",
					"supports":  gin.H{"dimensions": true},
					"tokenizer": "cl100k_base",
					"type":      "embeddings",
				},
				"id":                   "text-embedding-3-small",
				"name":                 "Embedding V3 small",
				"object":               "model",
				"version":              "text-embedding-3-small",
				"model_picker_enabled": false,
			},
			{
				"capabilities": gin.H{
					"family":    "text-embedding-3-small",
					"object":    "model_capabilities",
					"supports":  gin.H{"dimensions": true},
					"tokenizer": "cl100k_base",
					"type":      "embeddings",
				},
				"id":                   "text-embedding-3-small-inference",
				"name":                 "Embedding V3 small (Inference)",
				"object":               "model",
				"version":              "text-embedding-3-small",
				"model_picker_enabled": false,
			},
		},
		"object": "list",
	})
}

func CloseIO(c io.Closer) {
	err := c.Close()
	if nil != err {
		log.Println(err)
	}
}

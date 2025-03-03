package copilot

import (
	_ "embed"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
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
					"family": "gpt-3.5-turbo",
					"limits": gin.H{
						"max_context_window_tokens": 16384,
						"max_output_tokens":         4096,
						"max_prompt_tokens":         12288,
					},
					"object":    "model_capabilities",
					"supports":  gin.H{"streaming": true, "tool_calls": true},
					"tokenizer": "cl100k_base",
					"type":      "chat",
				},
				"id":                   "gpt-3.5-turbo",
				"model_picker_enabled": false,
				"name":                 "GPT 3.5 Turbo",
				"object":               "model",
				"preview":              false,
				"vendor":               "Azure OpenAI",
				"version":              "gpt-3.5-turbo-0613",
			},
			{
				"capabilities": gin.H{
					"family": "gpt-3.5-turbo",
					"limits": gin.H{
						"max_context_window_tokens": 16384,
						"max_output_tokens":         4096,
						"max_prompt_tokens":         12288,
					},
					"object":    "model_capabilities",
					"supports":  gin.H{"streaming": true, "tool_calls": true},
					"tokenizer": "cl100k_base",
					"type":      "chat",
				},
				"id":                   "gpt-3.5-turbo-0613",
				"model_picker_enabled": false,
				"name":                 "GPT 3.5 Turbo",
				"object":               "model",
				"preview":              false,
				"vendor":               "Azure OpenAI",
				"version":              "gpt-3.5-turbo-0613",
			},
			{
				"capabilities": gin.H{
					"family": "gpt-4",
					"limits": gin.H{
						"max_context_window_tokens": 32768,
						"max_output_tokens":         4096,
						"max_prompt_tokens":         32768,
					},
					"object":    "model_capabilities",
					"supports":  gin.H{"streaming": true, "tool_calls": true},
					"tokenizer": "cl100k_base",
					"type":      "chat",
				},
				"id":                   "gpt-4",
				"model_picker_enabled": false,
				"name":                 "GPT 4",
				"object":               "model",
				"preview":              false,
				"vendor":               "Azure OpenAI",
				"version":              "gpt-4-0613",
			},
			{
				"capabilities": gin.H{
					"family": "gpt-4",
					"limits": gin.H{
						"max_context_window_tokens": 32768,
						"max_output_tokens":         4096,
						"max_prompt_tokens":         32768,
					},
					"object":    "model_capabilities",
					"supports":  gin.H{"streaming": true, "tool_calls": true},
					"tokenizer": "cl100k_base",
					"type":      "chat",
				},
				"id":                   "gpt-4-0613",
				"model_picker_enabled": false,
				"name":                 "GPT 4",
				"object":               "model",
				"preview":              false,
				"vendor":               "Azure OpenAI",
				"version":              "gpt-4-0613",
			},
			{
				"capabilities": gin.H{
					"family": "gpt-4o",
					"limits": gin.H{
						"max_context_window_tokens": 128000,
						"max_output_tokens":         4096,
						"max_prompt_tokens":         64000,
						"vision": gin.H{
							"max_prompt_image_size": 3145728,
							"max_prompt_images":     1,
							"supported_media_types": []string{
								"image/jpeg",
								"image/png",
								"image/webp",
								"image/gif",
							},
						},
					},
					"object":    "model_capabilities",
					"supports":  gin.H{"parallel_tool_calls": true, "streaming": true, "tool_calls": true},
					"tokenizer": "o200k_base",
					"type":      "chat",
				},
				"id":                   "gpt-4o",
				"model_picker_enabled": true,
				"name":                 "GPT-4o",
				"object":               "model",
				"preview":              false,
				"vendor":               "Azure OpenAI",
				"version":              "gpt-4o-2024-05-13",
			},
			{
				"capabilities": gin.H{
					"family": "gpt-4o",
					"limits": gin.H{
						"max_context_window_tokens": 128000,
						"max_output_tokens":         4096,
						"max_prompt_tokens":         64000,
						"vision": gin.H{
							"max_prompt_image_size": 3145728,
							"max_prompt_images":     1,
							"supported_media_types": []string{
								"image/jpeg",
								"image/png",
								"image/webp",
								"image/gif",
							},
						},
					},
					"object":    "model_capabilities",
					"supports":  gin.H{"parallel_tool_calls": true, "streaming": true, "tool_calls": true},
					"tokenizer": "o200k_base",
					"type":      "chat",
				},
				"id":                   "gpt-4o-2024-05-13",
				"model_picker_enabled": false,
				"name":                 "GPT-4o",
				"object":               "model",
				"preview":              false,
				"vendor":               "Azure OpenAI",
				"version":              "gpt-4o-2024-05-13",
			},
			{
				"capabilities": gin.H{
					"family": "gpt-4o",
					"limits": gin.H{
						"max_context_window_tokens": 128000,
						"max_output_tokens":         4096,
						"max_prompt_tokens":         64000,
					},
					"object":    "model_capabilities",
					"supports":  gin.H{"parallel_tool_calls": true, "streaming": true, "tool_calls": true},
					"tokenizer": "o200k_base",
					"type":      "chat",
				},
				"id":                   "gpt-4-o-preview",
				"model_picker_enabled": false,
				"name":                 "GPT-4o",
				"object":               "model",
				"preview":              false,
				"vendor":               "Azure OpenAI",
				"version":              "gpt-4o-2024-05-13",
			},
			{
				"capabilities": gin.H{
					"family": "gpt-4o",
					"limits": gin.H{
						"max_context_window_tokens": 128000,
						"max_output_tokens":         16384,
						"max_prompt_tokens":         64000,
					},
					"object":    "model_capabilities",
					"supports":  gin.H{"parallel_tool_calls": true, "streaming": true, "tool_calls": true},
					"tokenizer": "o200k_base",
					"type":      "chat",
				},
				"id":                   "gpt-4o-2024-08-06",
				"model_picker_enabled": false,
				"name":                 "GPT-4o",
				"object":               "model",
				"preview":              false,
				"vendor":               "Azure OpenAI",
				"version":              "gpt-4o-2024-08-06",
			},
			{
				"capabilities": gin.H{
					"family": "text-embedding-ada-002",
					"limits": gin.H{
						"max_inputs": 256,
					},
					"object":    "model_capabilities",
					"supports":  gin.H{},
					"tokenizer": "cl100k_base",
					"type":      "embeddings",
				},
				"id":                   "text-embedding-ada-002",
				"model_picker_enabled": false,
				"name":                 "Embedding V2 Ada",
				"object":               "model",
				"preview":              false,
				"vendor":               "Azure OpenAI",
				"version":              "text-embedding-ada-002",
			},
			{
				"capabilities": gin.H{
					"family": "text-embedding-3-small",
					"limits": gin.H{
						"max_inputs": 512,
					},
					"object":    "model_capabilities",
					"supports":  gin.H{"dimensions": true},
					"tokenizer": "cl100k_base",
					"type":      "embeddings",
				},
				"id":                   "text-embedding-3-small",
				"model_picker_enabled": false,
				"name":                 "Embedding V3 small",
				"object":               "model",
				"preview":              false,
				"vendor":               "Azure OpenAI",
				"version":              "text-embedding-3-small",
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
				"model_picker_enabled": false,
				"name":                 "Embedding V3 small (Inference)",
				"object":               "model",
				"preview":              false,
				"vendor":               "Azure OpenAI",
				"version":              "text-embedding-3-small",
			},
			{
				"capabilities": gin.H{
					"family": "gpt-4o-mini",
					"limits": gin.H{
						"max_context_window_tokens": 128000,
						"max_output_tokens":         4096,
						"max_prompt_tokens":         12288,
					},
					"object":    "model_capabilities",
					"supports":  gin.H{"parallel_tool_calls": true, "streaming": true, "tool_calls": true},
					"tokenizer": "o200k_base",
					"type":      "chat",
				},
				"id":                   "gpt-4o-mini",
				"model_picker_enabled": false,
				"name":                 "GPT-4o mini",
				"object":               "model",
				"preview":              false,
				"vendor":               "Azure OpenAI",
				"version":              "gpt-4o-mini-2024-07-18",
			},
			{
				"capabilities": gin.H{
					"family": "gpt-4o-mini",
					"limits": gin.H{
						"max_context_window_tokens": 128000,
						"max_output_tokens":         4096,
						"max_prompt_tokens":         12288,
					},
					"object":    "model_capabilities",
					"supports":  gin.H{"parallel_tool_calls": true, "streaming": true, "tool_calls": true},
					"tokenizer": "o200k_base",
					"type":      "chat",
				},
				"id":                   "gpt-4o-mini-2024-07-18",
				"model_picker_enabled": false,
				"name":                 "GPT-4o mini",
				"object":               "model",
				"preview":              false,
				"vendor":               "Azure OpenAI",
				"version":              "gpt-4o-mini-2024-07-18",
			},
			{
				"capabilities": gin.H{
					"family": "o1-ga",
					"limits": gin.H{
						"max_context_window_tokens": 200000,
						"max_prompt_tokens":         20000,
					},
					"object":    "model_capabilities",
					"supports":  gin.H{"structured_outputs": true, "tool_calls": true},
					"tokenizer": "o200k_base",
					"type":      "chat",
				},
				"id":                   "o1",
				"model_picker_enabled": true,
				"name":                 "o1 (Preview)",
				"object":               "model",
				"preview":              true,
				"vendor":               "Azure OpenAI",
				"version":              "o1-2024-12-17",
			},
			{
				"capabilities": gin.H{
					"family": "o1-ga",
					"limits": gin.H{
						"max_context_window_tokens": 200000,
						"max_prompt_tokens":         20000,
					},
					"object":    "model_capabilities",
					"supports":  gin.H{"structured_outputs": true, "tool_calls": true},
					"tokenizer": "o200k_base",
					"type":      "chat",
				},
				"id":                   "o1-2024-12-17",
				"model_picker_enabled": false,
				"name":                 "o1 (Preview)",
				"object":               "model",
				"preview":              true,
				"vendor":               "Azure OpenAI",
				"version":              "o1-2024-12-17",
			},
			{
				"capabilities": gin.H{
					"family": "o3-mini",
					"limits": gin.H{
						"max_context_window_tokens": 200000,
						"max_output_tokens":         100000,
						"max_prompt_tokens":         64000,
					},
					"object":    "model_capabilities",
					"supports":  gin.H{"streaming": true, "structured_outputs": true, "tool_calls": true},
					"tokenizer": "o200k_base",
					"type":      "chat",
				},
				"id":                   "o3-mini",
				"model_picker_enabled": true,
				"name":                 "o3-mini (Preview)",
				"object":               "model",
				"preview":              true,
				"vendor":               "Azure OpenAI",
				"version":              "o3-mini-2025-01-31",
			},
			{
				"capabilities": gin.H{
					"family": "o3-mini",
					"limits": gin.H{
						"max_context_window_tokens": 200000,
						"max_output_tokens":         100000,
						"max_prompt_tokens":         64000,
					},
					"object":    "model_capabilities",
					"supports":  gin.H{"streaming": true, "structured_outputs": true, "tool_calls": true},
					"tokenizer": "o200k_base",
					"type":      "chat",
				},
				"id":                   "o3-mini-2025-01-31",
				"model_picker_enabled": false,
				"name":                 "o3-mini (Preview)",
				"object":               "model",
				"preview":              true,
				"vendor":               "Azure OpenAI",
				"version":              "o3-mini-2025-01-31",
			},
			{
				"capabilities": gin.H{
					"family": "o3-mini",
					"limits": gin.H{
						"max_context_window_tokens": 200000,
						"max_output_tokens":         100000,
						"max_prompt_tokens":         64000,
					},
					"object":    "model_capabilities",
					"supports":  gin.H{"streaming": true, "structured_outputs": true, "tool_calls": true},
					"tokenizer": "o200k_base",
					"type":      "chat",
				},
				"id":                   "o3-mini-paygo",
				"model_picker_enabled": false,
				"name":                 "o3-mini (Preview)",
				"object":               "model",
				"preview":              true,
				"vendor":               "Azure OpenAI",
				"version":              "o3-mini-paygo",
			},
			{
				"capabilities": gin.H{
					"family": "claude-3.5-sonnet",
					"limits": gin.H{
						"max_context_window_tokens": 90000,
						"max_output_tokens":         8192,
						"max_prompt_tokens":         90000,
						"vision": gin.H{
							"max_prompt_image_size": 3145728,
							"max_prompt_images":     1,
							"supported_media_types": []string{
								"image/jpeg",
								"image/png",
								"image/gif",
								"image/webp",
							},
						},
					},
					"object":    "model_capabilities",
					"supports":  gin.H{"parallel_tool_calls": true, "streaming": true, "tool_calls": true},
					"tokenizer": "o200k_base",
					"type":      "chat",
				},
				"id":                   "claude-3.5-sonnet",
				"model_picker_enabled": true,
				"name":                 "Claude 3.5 Sonnet (Preview)",
				"object":               "model",
				"policy": gin.H{
					"state": "enabled",
					"terms": "Enable access to the latest Claude 3.5 Sonnet model from Anthropic. [Learn more about how GitHub Copilot serves Claude 3.5 Sonnet](https://docs.github.com/copilot/using-github-copilot/using-claude-sonnet-in-github-copilot).",
				},
				"preview": true,
				"vendor":  "Anthropic",
				"version": "claude-3.5-sonnet",
			},
			{
				"capabilities": gin.H{
					"family": "claude-3.7-sonnet",
					"limits": gin.H{
						"max_context_window_tokens": 200000,
						"max_output_tokens":         8192,
						"max_prompt_tokens":         90000,
					},
					"object":    "model_capabilities",
					"supports":  gin.H{"parallel_tool_calls": true, "streaming": true, "tool_calls": true},
					"tokenizer": "o200k_base",
					"type":      "chat",
				},
				"id":                   "claude-3.7-sonnet",
				"model_picker_enabled": true,
				"name":                 "Claude 3.7 Sonnet (Preview)",
				"object":               "model",
				"policy": gin.H{
					"state": "enabled",
					"terms": "Enable access to the latest Claude 3.7 Sonnet model from Anthropic. [Learn more about how GitHub Copilot serves Claude 3.7 Sonnet](https://docs.github.com/copilot/using-github-copilot/using-claude-sonnet-in-github-copilot).",
				},
				"preview": true,
				"vendor":  "Anthropic",
				"version": "claude-3.7-sonnet",
			},
			{
				"capabilities": gin.H{
					"family": "claude-3.7-sonnet-thought",
					"limits": gin.H{
						"max_context_window_tokens": 200000,
						"max_output_tokens":         8192,
						"max_prompt_tokens":         90000,
					},
					"object":    "model_capabilities",
					"supports":  gin.H{"streaming": true},
					"tokenizer": "o200k_base",
					"type":      "chat",
				},
				"id":                   "claude-3.7-sonnet-thought",
				"model_picker_enabled": true,
				"name":                 "Claude 3.7 Sonnet Thinking (Preview)",
				"object":               "model",
				"policy": gin.H{
					"state": "enabled",
					"terms": "Enable access to the latest Claude 3.7 Sonnet model from Anthropic. [Learn more about how GitHub Copilot serves Claude 3.7 Sonnet](https://docs.github.com/copilot/using-github-copilot/using-claude-sonnet-in-github-copilot).",
				},
				"preview": true,
				"vendor":  "Anthropic",
				"version": "claude-3.7-sonnet-thought",
			},
			{
				"capabilities": gin.H{
					"family": "gemini-2.0-flash",
					"limits": gin.H{
						"max_context_window_tokens": 1000000,
						"max_output_tokens":         8192,
						"max_prompt_tokens":         128000,
						"vision": gin.H{
							"max_prompt_image_size": 3145728,
							"max_prompt_images":     1,
							"supported_media_types": []string{
								"image/jpeg",
								"image/png",
								"image/webp",
								"image/heic",
								"image/heif",
							},
						},
					},
					"object":    "model_capabilities",
					"supports":  gin.H{"streaming": true},
					"tokenizer": "o200k_base",
					"type":      "chat",
				},
				"id":                   "gemini-2.0-flash-001",
				"model_picker_enabled": true,
				"name":                 "Gemini 2.0 Flash (Preview)",
				"object":               "model",
				"policy": gin.H{
					"state": "enabled",
					"terms": "Enable access to the latest Gemini models from Google. [Learn more about how GitHub Copilot serves Gemini 2.0 Flash](https://docs.github.com/en/copilot/using-github-copilot/ai-models/using-gemini-flash-in-github-copilot).",
				},
				"preview": true,
				"vendor":  "Google",
				"version": "gemini-2.0-flash-001",
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

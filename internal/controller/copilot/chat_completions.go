package copilot

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// ChatCompletions chat对话接口
func ChatCompletions(c *gin.Context) {
	ctx := c.Request.Context()

	// 添加响应头, 解决vscode校验github所属问题
	requestID := uuid.Must(uuid.NewV4()).String()
	c.Header("x-github-request-id", requestID)

	body, err := io.ReadAll(c.Request.Body)
	if nil != err {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	apiModelName := gjson.GetBytes(body, "model").String()
	// 默认设置的对话模型
	envModelName := os.Getenv("CHAT_API_MODEL_NAME")
	// 默认设置的对话请求地址
	chatAPIURL := os.Getenv("CHAT_API_BASE")
	// 默认设置的对话模型key
	apiKey := os.Getenv("CHAT_API_KEY")

	// 轻量模型直接走代码补全接口, 节约成本
	if strings.Contains(apiModelName, os.Getenv("LIGHTWEIGHT_MODEL")) {
		envModelName = os.Getenv("CODEX_API_MODEL_NAME")
		codexAPIURL := os.Getenv("CODEX_API_BASE")
		chatAPIURL = strings.Replace(codexAPIURL, "/v1/completions", "/v1/chat/completions", 1)
		apiKey = os.Getenv("CODEX_API_KEY")
	}

	c.Header("Content-Type", "text/event-stream")

	body, _ = sjson.SetBytes(body, "model", envModelName)
	body, _ = sjson.SetBytes(body, "stream", true) // 强制流式输出

	if !gjson.GetBytes(body, "function_call").Exists() {
		messages := gjson.GetBytes(body, "messages").Array()
		for i, msg := range messages {
			toolCalls := msg.Get("tool_calls").Array()
			if len(toolCalls) == 0 {
				body, _ = sjson.DeleteBytes(body, fmt.Sprintf("messages.%d.tool_calls", i))
			}
		}
		lastIndex := len(messages) - 1
		chatLocale := os.Getenv("CHAT_LOCALE")
		if chatLocale != "" && !strings.Contains(messages[lastIndex].Get("content").String(), "Respond in the following locale") {
			body, _ = sjson.SetBytes(body, "messages."+strconv.Itoa(lastIndex)+".content", messages[lastIndex].Get("content").String()+"Respond in the following locale: "+chatLocale+".")
		}
	}

	body, _ = sjson.DeleteBytes(body, "intent")
	body, _ = sjson.DeleteBytes(body, "intent_threshold")
	body, _ = sjson.DeleteBytes(body, "intent_content")
	body, _ = sjson.DeleteBytes(body, "logprobs") // #IBZYCA

	// 是否支持使用工具, 避免模型不支持相关功能报错
	chatUseTools, _ := strconv.ParseBool(os.Getenv("CHAT_USE_TOOLS"))
	if !chatUseTools {
		body, _ = sjson.DeleteBytes(body, "tools")
		body, _ = sjson.DeleteBytes(body, "tool_call")
		body, _ = sjson.DeleteBytes(body, "functions")
		body, _ = sjson.DeleteBytes(body, "function_call")
		body, _ = sjson.DeleteBytes(body, "tool_choice")
	}

	ChatMaxTokens, _ := strconv.Atoi(os.Getenv("CHAT_MAX_TOKENS"))
	if int(gjson.GetBytes(body, "max_tokens").Int()) > ChatMaxTokens {
		body, _ = sjson.SetBytes(body, "max_tokens", ChatMaxTokens)
	}

	if gjson.GetBytes(body, "n").Int() > 1 {
		body, _ = sjson.SetBytes(body, "n", 1)
	}

	messages := gjson.GetBytes(body, "messages").Array()
	userAgent := c.GetHeader("User-Agent")

	// 拦截处理vscode对话首次预处理请求, 减少等待时间
	firstRole := gjson.GetBytes(body, "messages.0.role").String()
	firstContent := gjson.GetBytes(body, "messages.0.content").String()
	if strings.Contains(firstRole, "system") && strings.Contains(firstContent, "You are a helpful AI programming assistant to a user") &&
		!strings.Contains(firstContent, "If you cannot choose just one category, or if none of the categories seem like they would provide the user with a better result, you must always respond with") &&
		!gjson.GetBytes(body, "tool_choice").Exists() {
		_, _ = c.Writer.WriteString("data: [DONE]\n\n")
		c.Writer.Flush()
		return
	}

	// vs2022客户端的兼容处理
	if strings.Contains(userAgent, "VSCopilotClient") {
		lastMessage := messages[len(messages)-1]
		messageRole := lastMessage.Get("role").String()
		messageContent := lastMessage.Get("content").String()
		if strings.Contains(firstRole, "system") && strings.Contains(firstContent, "You are an AI programming assistant") {
			vs2022FirstChatTemplate(c)
			return
		}
		if messageRole == "user" && messageContent == "Write a short one-sentence question that I can ask that naturally follows from the previous few questions and answers. It should not ask a question which is already answered in the conversation. It should be a question that you are capable of answering. Reply with only the text of the question and nothing else." {
			_, _ = c.Writer.WriteString("data: [DONE]\n\n")
			c.Writer.Flush()
			return
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, chatAPIURL, io.NopCloser(bytes.NewBuffer(body)))
	if nil != err {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	httpClientTimeout, _ := time.ParseDuration(os.Getenv("HTTP_CLIENT_TIMEOUT") + "s")
	client := &http.Client{
		Timeout: httpClientTimeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	resp, err := client.Do(req)
	if nil != err {
		if errors.Is(err, context.Canceled) {
			c.AbortWithStatus(http.StatusRequestTimeout)
			return
		}

		log.Println("request conversation failed:", err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	defer CloseIO(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Println("request completions failed:", string(body))

		resp.Body = io.NopCloser(bytes.NewBuffer(body))
	}

	c.Status(resp.StatusCode)
	_, _ = io.Copy(c.Writer, resp.Body)
}

// vs2022FirstChatTemplate is a template for the first chat completion response
func vs2022FirstChatTemplate(c *gin.Context) {
	fixedOutput := `data: {"id":"f6202f6f-9d13-4518-b34f-65e945b0a1a2","object":"chat.completion.chunk","model":"gpt-4o-mini-2024-07-18","created":1734752124,"choices":[{"index":0,"delta":{"role":"assistant","content":""},"finish_reason":null}]}

data: {"id":"b2ab39cb-9a84-4006-b470-93a5965c6d69","object":"chat.completion.chunk","model":"gpt-4o-mini-2024-07-18","created":1734752124,"choices":[{"index":0,"delta":{"role":"assistant","content":""},"finish_reason":null}]}

data: {"id":"df5f9ce7-b653-4ffb-8d92-e21856ce1ffc","object":"chat.completion.chunk","model":"gpt-4o-mini-2024-07-18","created":1734752124,"choices":[{"index":0,"delta":{"role":"assistant","content":"Explain"},"finish_reason":null}]}

data: {"id":"fb58d66e-bb16-43f2-8470-2de0c8662533","object":"chat.completion.chunk","model":"gpt-4o-mini-2024-07-18","created":1734752124,"choices":[{"index":0,"delta":{"role":"assistant","content":""},"finish_reason":null}]}

data: {"id":"22ea16e2-766f-4b10-84d0-68399abc9181","object":"chat.completion.chunk","model":"gpt-4o-mini-2024-07-18","created":1734752124,"choices":[{"index":0,"delta":{"role":"assistant","content":""},"finish_reason":"stop"}]}

data: [DONE]

`
	_, _ = c.Writer.WriteString(fixedOutput)
	c.Writer.Flush()
}

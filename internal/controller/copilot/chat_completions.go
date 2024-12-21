package copilot

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
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

	body, err := io.ReadAll(c.Request.Body)
	if nil != err {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	envModelName := os.Getenv("CHAT_API_MODEL_NAME")
	c.Header("Content-Type", "text/event-stream")
	body, _ = sjson.SetBytes(body, "model", envModelName)

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

	if !strings.HasPrefix(envModelName, "gpt-") {
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
	// vs2022客户端的兼容处理
	if strings.Contains(userAgent, "VSCopilotClient") {
		lastMessage := messages[len(messages)-1]
		messageRole := lastMessage.Get("role").String()
		messageContent := lastMessage.Get("content").String()
		firstRole := gjson.GetBytes(body, "messages.0.role").String()
		firstContent := gjson.GetBytes(body, "messages.0.content").String()
		if strings.Contains(firstRole, "system") && firstContent == "You are an AI programming assistant." {
			// todo: 解决多轮对话场景, 但是在代码选中后右键选择解释代码会报错, 解决方法是点击一下"在聊天窗口中继续"即可.
			vs2022FirstChatTemplate(c)
			return
		}
		if messageRole == "user" && messageContent == "Write a short one-sentence question that I can ask that naturally follows from the previous few questions and answers. It should not ask a question which is already answered in the conversation. It should be a question that you are capable of answering. Reply with only the text of the question and nothing else." {
			_, _ = c.Writer.WriteString("data: [DONE]\n\n")
			c.Writer.Flush()
			return
		}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, os.Getenv("CHAT_API_BASE"), io.NopCloser(bytes.NewBuffer(body)))
	if nil != err {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("CHAT_API_KEY"))
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

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

func chatCompletions(c *gin.Context) {
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

	client := &http.Client{
		Timeout: 30 * time.Second,
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
	defer closeIO(resp.Body)

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
	fixedOutput := `data: {"choices":[],"created":0,"id":"","prompt_filter_results":[{"content_filter_results":{"error":{"code":"","message":""},"hate":{"filtered":false,"severity":"safe"},"self_harm":{"filtered":false,"severity":"safe"},"sexual":{"filtered":false,"severity":"safe"},"violence":{"filtered":false,"severity":"safe"}},"prompt_index":0}]}

data: {"choices":[{"index":0,"content_filter_offsets":{"check_offset":2570,"start_offset":2570,"end_offset":2576},"content_filter_results":{"error":{"code":"","message":""},"hate":{"filtered":false,"severity":"safe"},"self_harm":{"filtered":false,"severity":"safe"},"sexual":{"filtered":false,"severity":"safe"},"violence":{"filtered":false,"severity":"safe"}},"delta":{"content":"","role":"assistant"}}],"created":1727587479,"id":"chatcmpl-ACgi7bsB68mFrBhjvCcD95lHwClfN","model":"gpt-3.5-turbo-0613"}

data: {"choices":[{"index":0,"content_filter_offsets":{"check_offset":2570,"start_offset":2570,"end_offset":2576},"content_filter_results":{"error":{"code":"","message":""},"hate":{"filtered":false,"severity":"safe"},"self_harm":{"filtered":false,"severity":"safe"},"sexual":{"filtered":false,"severity":"safe"},"violence":{"filtered":false,"severity":"safe"}},"delta":{"content":"Answer"}}],"created":1727587479,"id":"chatcmpl-ACgi7bsB68mFrBhjvCcD95lHwClfN","model":"gpt-3.5-turbo-0613"}

data: {"choices":[{"finish_reason":"stop","index":0,"content_filter_offsets":{"check_offset":2570,"start_offset":2570,"end_offset":2576},"content_filter_results":{"error":{"code":"","message":""},"hate":{"filtered":false,"severity":"safe"},"self_harm":{"filtered":false,"severity":"safe"},"sexual":{"filtered":false,"severity":"safe"},"violence":{"filtered":false,"severity":"safe"}},"delta":{"content":null}}],"created":1727587479,"id":"chatcmpl-ACgi7bsB68mFrBhjvCcD95lHwClfN","usage":{"completion_tokens":1,"prompt_tokens":519,"total_tokens":520},"model":"gpt-3.5-turbo-0613"}

data: {"choices":[{"index":1,"content_filter_offsets":{"check_offset":2570,"start_offset":2570,"end_offset":2576},"content_filter_results":{"error":{"code":"","message":""},"hate":{"filtered":false,"severity":"safe"},"self_harm":{"filtered":false,"severity":"safe"},"sexual":{"filtered":false,"severity":"safe"},"violence":{"filtered":false,"severity":"safe"}},"delta":{"content":"","role":"assistant"}}],"created":1727587479,"id":"chatcmpl-ACgi7bsB68mFrBhjvCcD95lHwClfN","model":"gpt-3.5-turbo-0613"}

data: {"choices":[{"index":1,"content_filter_offsets":{"check_offset":2570,"start_offset":2570,"end_offset":2576},"content_filter_results":{"error":{"code":"","message":""},"hate":{"filtered":false,"severity":"safe"},"self_harm":{"filtered":false,"severity":"safe"},"sexual":{"filtered":false,"severity":"safe"},"violence":{"filtered":false,"severity":"safe"}},"delta":{"content":"Answer"}}],"created":1727587479,"id":"chatcmpl-ACgi7bsB68mFrBhjvCcD95lHwClfN","model":"gpt-3.5-turbo-0613"}

data: {"choices":[{"finish_reason":"stop","index":1,"content_filter_offsets":{"check_offset":2570,"start_offset":2570,"end_offset":2576},"content_filter_results":{"error":{"code":"","message":""},"hate":{"filtered":false,"severity":"safe"},"self_harm":{"filtered":false,"severity":"safe"},"sexual":{"filtered":false,"severity":"safe"},"violence":{"filtered":false,"severity":"safe"}},"delta":{"content":null}}],"created":1727587479,"id":"chatcmpl-ACgi7bsB68mFrBhjvCcD95lHwClfN","usage":{"completion_tokens":2,"prompt_tokens":519,"total_tokens":521},"model":"gpt-3.5-turbo-0613"}

data: {"choices":[{"index":2,"content_filter_offsets":{"check_offset":2570,"start_offset":2570,"end_offset":2576},"content_filter_results":{"error":{"code":"","message":""},"hate":{"filtered":false,"severity":"safe"},"self_harm":{"filtered":false,"severity":"safe"},"sexual":{"filtered":false,"severity":"safe"},"violence":{"filtered":false,"severity":"safe"}},"delta":{"content":"","role":"assistant"}}],"created":1727587479,"id":"chatcmpl-ACgi7bsB68mFrBhjvCcD95lHwClfN","model":"gpt-3.5-turbo-0613"}

data: {"choices":[{"index":2,"content_filter_offsets":{"check_offset":2570,"start_offset":2570,"end_offset":2576},"content_filter_results":{"error":{"code":"","message":""},"hate":{"filtered":false,"severity":"safe"},"self_harm":{"filtered":false,"severity":"safe"},"sexual":{"filtered":false,"severity":"safe"},"violence":{"filtered":false,"severity":"safe"}},"delta":{"content":"Answer"}}],"created":1727587479,"id":"chatcmpl-ACgi7bsB68mFrBhjvCcD95lHwClfN","model":"gpt-3.5-turbo-0613"}

data: {"choices":[{"finish_reason":"stop","index":2,"content_filter_offsets":{"check_offset":2570,"start_offset":2570,"end_offset":2576},"content_filter_results":{"error":{"code":"","message":""},"hate":{"filtered":false,"severity":"safe"},"self_harm":{"filtered":false,"severity":"safe"},"sexual":{"filtered":false,"severity":"safe"},"violence":{"filtered":false,"severity":"safe"}},"delta":{"content":null}}],"created":1727587479,"id":"chatcmpl-ACgi7bsB68mFrBhjvCcD95lHwClfN","usage":{"completion_tokens":3,"prompt_tokens":519,"total_tokens":522},"model":"gpt-3.5-turbo-0613"}

data: {"choices":[{"index":3,"content_filter_offsets":{"check_offset":2570,"start_offset":2570,"end_offset":2576},"content_filter_results":{"error":{"code":"","message":""},"hate":{"filtered":false,"severity":"safe"},"self_harm":{"filtered":false,"severity":"safe"},"sexual":{"filtered":false,"severity":"safe"},"violence":{"filtered":false,"severity":"safe"}},"delta":{"content":"","role":"assistant"}}],"created":1727587479,"id":"chatcmpl-ACgi7bsB68mFrBhjvCcD95lHwClfN","model":"gpt-3.5-turbo-0613"}

data: {"choices":[{"index":3,"content_filter_offsets":{"check_offset":2570,"start_offset":2570,"end_offset":2576},"content_filter_results":{"error":{"code":"","message":""},"hate":{"filtered":false,"severity":"safe"},"self_harm":{"filtered":false,"severity":"safe"},"sexual":{"filtered":false,"severity":"safe"},"violence":{"filtered":false,"severity":"safe"}},"delta":{"content":"Answer"}}],"created":1727587479,"id":"chatcmpl-ACgi7bsB68mFrBhjvCcD95lHwClfN","model":"gpt-3.5-turbo-0613"}

data: {"choices":[{"finish_reason":"stop","index":3,"content_filter_offsets":{"check_offset":2570,"start_offset":2570,"end_offset":2576},"content_filter_results":{"error":{"code":"","message":""},"hate":{"filtered":false,"severity":"safe"},"self_harm":{"filtered":false,"severity":"safe"},"sexual":{"filtered":false,"severity":"safe"},"violence":{"filtered":false,"severity":"safe"}},"delta":{"content":null}}],"created":1727587479,"id":"chatcmpl-ACgi7bsB68mFrBhjvCcD95lHwClfN","usage":{"completion_tokens":4,"prompt_tokens":519,"total_tokens":523},"model":"gpt-3.5-turbo-0613"}

data: {"choices":[{"index":4,"content_filter_offsets":{"check_offset":2570,"start_offset":2570,"end_offset":2576},"content_filter_results":{"error":{"code":"","message":""},"hate":{"filtered":false,"severity":"safe"},"self_harm":{"filtered":false,"severity":"safe"},"sexual":{"filtered":false,"severity":"safe"},"violence":{"filtered":false,"severity":"safe"}},"delta":{"content":"","role":"assistant"}}],"created":1727587479,"id":"chatcmpl-ACgi7bsB68mFrBhjvCcD95lHwClfN","model":"gpt-3.5-turbo-0613"}

data: {"choices":[{"index":4,"content_filter_offsets":{"check_offset":2570,"start_offset":2570,"end_offset":2576},"content_filter_results":{"error":{"code":"","message":""},"hate":{"filtered":false,"severity":"safe"},"self_harm":{"filtered":false,"severity":"safe"},"sexual":{"filtered":false,"severity":"safe"},"violence":{"filtered":false,"severity":"safe"}},"delta":{"content":"Answer"}}],"created":1727587479,"id":"chatcmpl-ACgi7bsB68mFrBhjvCcD95lHwClfN","model":"gpt-3.5-turbo-0613"}

data: {"choices":[{"finish_reason":"stop","index":4,"content_filter_offsets":{"check_offset":2570,"start_offset":2570,"end_offset":2576},"content_filter_results":{"error":{"code":"","message":""},"hate":{"filtered":false,"severity":"safe"},"self_harm":{"filtered":false,"severity":"safe"},"sexual":{"filtered":false,"severity":"safe"},"violence":{"filtered":false,"severity":"safe"}},"delta":{"content":null}}],"created":1727587479,"id":"chatcmpl-ACgi7bsB68mFrBhjvCcD95lHwClfN","usage":{"completion_tokens":5,"prompt_tokens":519,"total_tokens":524},"model":"gpt-3.5-turbo-0613"}

data: {"choices":[{"index":5,"content_filter_offsets":{"check_offset":2570,"start_offset":2570,"end_offset":2576},"content_filter_results":{"error":{"code":"","message":""},"hate":{"filtered":false,"severity":"safe"},"self_harm":{"filtered":false,"severity":"safe"},"sexual":{"filtered":false,"severity":"safe"},"violence":{"filtered":false,"severity":"safe"}},"delta":{"content":"","role":"assistant"}}],"created":1727587479,"id":"chatcmpl-ACgi7bsB68mFrBhjvCcD95lHwClfN","model":"gpt-3.5-turbo-0613"}

data: {"choices":[{"index":5,"content_filter_offsets":{"check_offset":2570,"start_offset":2570,"end_offset":2576},"content_filter_results":{"error":{"code":"","message":""},"hate":{"filtered":false,"severity":"safe"},"self_harm":{"filtered":false,"severity":"safe"},"sexual":{"filtered":false,"severity":"safe"},"violence":{"filtered":false,"severity":"safe"}},"delta":{"content":"Answer"}}],"created":1727587479,"id":"chatcmpl-ACgi7bsB68mFrBhjvCcD95lHwClfN","model":"gpt-3.5-turbo-0613"}

data: {"choices":[{"finish_reason":"stop","index":5,"content_filter_offsets":{"check_offset":2570,"start_offset":2570,"end_offset":2576},"content_filter_results":{"error":{"code":"","message":""},"hate":{"filtered":false,"severity":"safe"},"self_harm":{"filtered":false,"severity":"safe"},"sexual":{"filtered":false,"severity":"safe"},"violence":{"filtered":false,"severity":"safe"}},"delta":{"content":null}}],"created":1727587479,"id":"chatcmpl-ACgi7bsB68mFrBhjvCcD95lHwClfN","usage":{"completion_tokens":6,"prompt_tokens":519,"total_tokens":525},"model":"gpt-3.5-turbo-0613"}

data: {"choices":[{"index":6,"content_filter_offsets":{"check_offset":2570,"start_offset":2570,"end_offset":2576},"content_filter_results":{"error":{"code":"","message":""},"hate":{"filtered":false,"severity":"safe"},"self_harm":{"filtered":false,"severity":"safe"},"sexual":{"filtered":false,"severity":"safe"},"violence":{"filtered":false,"severity":"safe"}},"delta":{"content":"","role":"assistant"}}],"created":1727587479,"id":"chatcmpl-ACgi7bsB68mFrBhjvCcD95lHwClfN","model":"gpt-3.5-turbo-0613"}

data: {"choices":[{"index":6,"content_filter_offsets":{"check_offset":2570,"start_offset":2570,"end_offset":2576},"content_filter_results":{"error":{"code":"","message":""},"hate":{"filtered":false,"severity":"safe"},"self_harm":{"filtered":false,"severity":"safe"},"sexual":{"filtered":false,"severity":"safe"},"violence":{"filtered":false,"severity":"safe"}},"delta":{"content":"Answer"}}],"created":1727587479,"id":"chatcmpl-ACgi7bsB68mFrBhjvCcD95lHwClfN","model":"gpt-3.5-turbo-0613"}

data: {"choices":[{"finish_reason":"stop","index":6,"content_filter_offsets":{"check_offset":2570,"start_offset":2570,"end_offset":2576},"content_filter_results":{"error":{"code":"","message":""},"hate":{"filtered":false,"severity":"safe"},"self_harm":{"filtered":false,"severity":"safe"},"sexual":{"filtered":false,"severity":"safe"},"violence":{"filtered":false,"severity":"safe"}},"delta":{"content":null}}],"created":1727587479,"id":"chatcmpl-ACgi7bsB68mFrBhjvCcD95lHwClfN","usage":{"completion_tokens":7,"prompt_tokens":519,"total_tokens":526},"model":"gpt-3.5-turbo-0613"}

data: {"choices":[{"index":7,"content_filter_offsets":{"check_offset":2570,"start_offset":2570,"end_offset":2576},"content_filter_results":{"error":{"code":"","message":""},"hate":{"filtered":false,"severity":"safe"},"self_harm":{"filtered":false,"severity":"safe"},"sexual":{"filtered":false,"severity":"safe"},"violence":{"filtered":false,"severity":"safe"}},"delta":{"content":"","role":"assistant"}}],"created":1727587479,"id":"chatcmpl-ACgi7bsB68mFrBhjvCcD95lHwClfN","model":"gpt-3.5-turbo-0613"}

data: {"choices":[{"index":7,"content_filter_offsets":{"check_offset":2570,"start_offset":2570,"end_offset":2576},"content_filter_results":{"error":{"code":"","message":""},"hate":{"filtered":false,"severity":"safe"},"self_harm":{"filtered":false,"severity":"safe"},"sexual":{"filtered":false,"severity":"safe"},"violence":{"filtered":false,"severity":"safe"}},"delta":{"content":"Answer"}}],"created":1727587479,"id":"chatcmpl-ACgi7bsB68mFrBhjvCcD95lHwClfN","model":"gpt-3.5-turbo-0613"}

data: {"choices":[{"finish_reason":"stop","index":7,"content_filter_offsets":{"check_offset":2570,"start_offset":2570,"end_offset":2576},"content_filter_results":{"error":{"code":"","message":""},"hate":{"filtered":false,"severity":"safe"},"self_harm":{"filtered":false,"severity":"safe"},"sexual":{"filtered":false,"severity":"safe"},"violence":{"filtered":false,"severity":"safe"}},"delta":{"content":null}}],"created":1727587479,"id":"chatcmpl-ACgi7bsB68mFrBhjvCcD95lHwClfN","usage":{"completion_tokens":8,"prompt_tokens":519,"total_tokens":527},"model":"gpt-3.5-turbo-0613"}

data: {"choices":[{"index":8,"content_filter_offsets":{"check_offset":2570,"start_offset":2570,"end_offset":2576},"content_filter_results":{"error":{"code":"","message":""},"hate":{"filtered":false,"severity":"safe"},"self_harm":{"filtered":false,"severity":"safe"},"sexual":{"filtered":false,"severity":"safe"},"violence":{"filtered":false,"severity":"safe"}},"delta":{"content":"","role":"assistant"}}],"created":1727587479,"id":"chatcmpl-ACgi7bsB68mFrBhjvCcD95lHwClfN","model":"gpt-3.5-turbo-0613"}

data: {"choices":[{"index":8,"content_filter_offsets":{"check_offset":2570,"start_offset":2570,"end_offset":2576},"content_filter_results":{"error":{"code":"","message":""},"hate":{"filtered":false,"severity":"safe"},"self_harm":{"filtered":false,"severity":"safe"},"sexual":{"filtered":false,"severity":"safe"},"violence":{"filtered":false,"severity":"safe"}},"delta":{"content":"Answer"}}],"created":1727587479,"id":"chatcmpl-ACgi7bsB68mFrBhjvCcD95lHwClfN","model":"gpt-3.5-turbo-0613"}

data: {"choices":[{"finish_reason":"stop","index":8,"content_filter_offsets":{"check_offset":2570,"start_offset":2570,"end_offset":2576},"content_filter_results":{"error":{"code":"","message":""},"hate":{"filtered":false,"severity":"safe"},"self_harm":{"filtered":false,"severity":"safe"},"sexual":{"filtered":false,"severity":"safe"},"violence":{"filtered":false,"severity":"safe"}},"delta":{"content":null}}],"created":1727587479,"id":"chatcmpl-ACgi7bsB68mFrBhjvCcD95lHwClfN","usage":{"completion_tokens":9,"prompt_tokens":519,"total_tokens":528},"model":"gpt-3.5-turbo-0613"}

data: {"choices":[{"index":9,"content_filter_offsets":{"check_offset":2570,"start_offset":2570,"end_offset":2576},"content_filter_results":{"error":{"code":"","message":""},"hate":{"filtered":false,"severity":"safe"},"self_harm":{"filtered":false,"severity":"safe"},"sexual":{"filtered":false,"severity":"safe"},"violence":{"filtered":false,"severity":"safe"}},"delta":{"content":"","role":"assistant"}}],"created":1727587479,"id":"chatcmpl-ACgi7bsB68mFrBhjvCcD95lHwClfN","model":"gpt-3.5-turbo-0613"}

data: {"choices":[{"index":9,"content_filter_offsets":{"check_offset":2570,"start_offset":2570,"end_offset":2576},"content_filter_results":{"error":{"code":"","message":""},"hate":{"filtered":false,"severity":"safe"},"self_harm":{"filtered":false,"severity":"safe"},"sexual":{"filtered":false,"severity":"safe"},"violence":{"filtered":false,"severity":"safe"}},"delta":{"content":"Answer"}}],"created":1727587479,"id":"chatcmpl-ACgi7bsB68mFrBhjvCcD95lHwClfN","model":"gpt-3.5-turbo-0613"}

data: {"choices":[{"finish_reason":"stop","index":9,"content_filter_offsets":{"check_offset":2570,"start_offset":2570,"end_offset":2576},"content_filter_results":{"error":{"code":"","message":""},"hate":{"filtered":false,"severity":"safe"},"self_harm":{"filtered":false,"severity":"safe"},"sexual":{"filtered":false,"severity":"safe"},"violence":{"filtered":false,"severity":"safe"}},"delta":{"content":null}}],"created":1727587479,"id":"chatcmpl-ACgi7bsB68mFrBhjvCcD95lHwClfN","usage":{"completion_tokens":10,"prompt_tokens":519,"total_tokens":529},"model":"gpt-3.5-turbo-0613"}

data: [DONE]

`
	_, _ = c.Writer.WriteString(fixedOutput)
	c.Writer.Flush()
}

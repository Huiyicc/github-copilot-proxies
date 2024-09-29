package copilot

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func chatCompletions(c *gin.Context) {
	ctx := c.Request.Context()

	body, err := io.ReadAll(c.Request.Body)
	if nil != err {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	c.Header("Content-Type", "text/event-stream")
	body, _ = sjson.SetBytes(body, "model", os.Getenv("CHAT_API_MODEL_NAME"))
	body, _ = sjson.DeleteBytes(body, "intent")
	body, _ = sjson.DeleteBytes(body, "intent_threshold")
	body, _ = sjson.DeleteBytes(body, "intent_content")
	body, _ = sjson.DeleteBytes(body, "tools")
	body, _ = sjson.DeleteBytes(body, "tool_call")
	body, _ = sjson.DeleteBytes(body, "functions")
	body, _ = sjson.DeleteBytes(body, "function_call")
	body, _ = sjson.DeleteBytes(body, "tool_choice")

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

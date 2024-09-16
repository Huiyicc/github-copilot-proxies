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

func chatCompletionsV2(c *gin.Context) {
	ctx := c.Request.Context()

	body, err := io.ReadAll(c.Request.Body)
	if nil != err {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	defer c.Request.Body.Close()

	body, _ = sjson.SetBytes(body, "model", os.Getenv("CHAT_API_MODEL_NAME"))

	if !gjson.GetBytes(body, "function_call").Exists() {
		messages := gjson.GetBytes(body, "messages").Array()
		lastIndex := len(messages) - 1
		if !strings.Contains(messages[lastIndex].Get("content").String(), "Respond in the following locale") {
			body, _ = sjson.SetBytes(body, "messages."+strconv.Itoa(lastIndex)+".content", messages[lastIndex].Get("content").String()+"Respond in the following locale: "+os.Getenv("CHAT_LOCALE")+".")
		}
	}

	body, _ = sjson.DeleteBytes(body, "intent")
	body, _ = sjson.DeleteBytes(body, "intent_threshold")
	body, _ = sjson.DeleteBytes(body, "intent_content")
	body, _ = sjson.SetBytes(body, "stream", true)

	ChatMaxTokens, _ := strconv.Atoi(os.Getenv("CHAT_MAX_TOKENS"))
	if int(gjson.GetBytes(body, "max_tokens").Int()) > ChatMaxTokens {
		body, _ = sjson.SetBytes(body, "max_tokens", ChatMaxTokens)
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

	c.Status(resp.StatusCode)

	contentType := resp.Header.Get("Content-Type")
	if "" != contentType {
		c.Header("Content-Type", contentType)
	}

	pr, pw := io.Pipe() // 创建管道

	// 启动一个 goroutine 用于从 resp.Body 读取数据并写入到管道
	go func() {
		defer pw.Close()
		_, err := io.Copy(pw, resp.Body)
		if err != nil {
			log.Println("Error copying response body:", err)
		}
	}()

	// 从管道中读取数据并写入到客户端
	_, err = io.Copy(c.Writer, pr)
	if err != nil {
		log.Println("Error writing to client:", err)
	}
}

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
	"time"
)

// codeCompletions 代码补全
func codeCompletions(c *gin.Context) {
	ctx := c.Request.Context()
	debounceTime, _ := strconv.Atoi(os.Getenv("COPILOT_DEBOUNCE"))
	time.Sleep(time.Duration(debounceTime) * time.Millisecond)

	if ctx.Err() != nil {
		abortCodex(c, http.StatusRequestTimeout)
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if nil != err {
		abortCodex(c, http.StatusBadRequest)
		return
	}

	body = ConstructRequestBody(body)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, os.Getenv("CODEX_API_BASE"), io.NopCloser(bytes.NewBuffer(body)))
	if nil != err {
		abortCodex(c, http.StatusInternalServerError)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("CODEX_API_KEY"))

	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	resp, err := client.Do(req)
	if nil != err {
		if errors.Is(err, context.Canceled) {
			abortCodex(c, http.StatusRequestTimeout)
			return
		}

		log.Println("request completions failed:", err.Error())
		abortCodex(c, http.StatusInternalServerError)
		return
	}
	defer closeIO(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Println("request completions failed:", string(body))

		abortCodex(c, resp.StatusCode)
		return
	}

	c.Status(resp.StatusCode)

	contentType := resp.Header.Get("Content-Type")
	if "" != contentType {
		c.Header("Content-Type", contentType)
	}

	_, _ = io.Copy(c.Writer, resp.Body)
}

// ConstructRequestBody 重新构建请求体
func ConstructRequestBody(body []byte) []byte {
	body, _ = sjson.DeleteBytes(body, "extra")
	body, _ = sjson.DeleteBytes(body, "nwo")
	// 重写 model 参数值
	body, _ = sjson.SetBytes(body, "model", os.Getenv("CODEX_API_MODEL_NAME"))

	// 重写 max_tokens 参数值
	CodeMaxTokens, _ := strconv.Atoi(os.Getenv("CODEX_MAX_TOKENS"))
	if int(gjson.GetBytes(body, "max_tokens").Int()) > CodeMaxTokens {
		body, _ = sjson.SetBytes(body, "max_tokens", CodeMaxTokens)
	}

	if gjson.GetBytes(body, "n").Int() > 1 {
		body, _ = sjson.SetBytes(body, "n", 1)
	}

	temperature, _ := strconv.ParseFloat(os.Getenv("CODEX_TEMPERATURE"), 64)
	if temperature != -1 {
		// 重写 temperature 参数值
		body, _ = sjson.SetBytes(body, "temperature", temperature)
	}
	return body
}

func abortCodex(c *gin.Context, status int) {
	c.Header("Content-Type", "text/event-stream")

	c.String(status, "data: [DONE]\n")
	c.Abort()
}

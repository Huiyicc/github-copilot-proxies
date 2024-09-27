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
	// 重置参数值已符合环境变量配置
	envCodexModel := os.Getenv("CODEX_API_MODEL_NAME")
	body, _ = sjson.SetBytes(body, "model", envCodexModel)

	temperature, _ := strconv.ParseFloat(os.Getenv("CODEX_TEMPERATURE"), 64)
	if temperature != -1 {
		body, _ = sjson.SetBytes(body, "temperature", temperature)
	}

	codeMaxTokens, _ := strconv.Atoi(os.Getenv("CODEX_MAX_TOKENS"))

	if strings.Contains(envCodexModel, "stable-code") || strings.Contains(envCodexModel, "codegemma") {
		return constructWithStableCodeModel(body)
	}

	if strings.Contains(envCodexModel, "codellama") {
		return constructWithCodeLlamaModel(body)
	}

	if strings.Contains(envCodexModel, "qwen-coder-turbo") || strings.Contains(envCodexModel, "qwen-coder-turbo-latest") {
		return constructWithQwenCoderTurboModel(body, codeMaxTokens)
	}

	return constructWithDeepSeekModel(body, codeMaxTokens)
}

// constructWithDeepSeekModel 重写DeepSeek模型要求的请求体
func constructWithDeepSeekModel(body []byte, codeMaxTokens int) []byte {
	if int(gjson.GetBytes(body, "max_tokens").Int()) > codeMaxTokens {
		body, _ = sjson.SetBytes(body, "max_tokens", codeMaxTokens)
	}

	if gjson.GetBytes(body, "n").Int() > 1 {
		body, _ = sjson.SetBytes(body, "n", 1)
	}
	return body
}

// constructWithCodeLlamaModel 重写codeLlama模型要求的请求体
func constructWithCodeLlamaModel(body []byte) []byte {
	suffix := gjson.GetBytes(body, "suffix")
	prompt := gjson.GetBytes(body, "prompt")
	content := fmt.Sprintf("<PRE> %s <SUF> %s <MID>", prompt, suffix)

	return constructWithChatModel(body, content)
}

// constructWithStableCodeModel 重写StableCode模型要求的请求体
func constructWithStableCodeModel(body []byte) []byte {
	suffix := gjson.GetBytes(body, "suffix")
	prompt := gjson.GetBytes(body, "prompt")
	content := fmt.Sprintf("<fim_prefix>%s<fim_suffix>%s<fim_middle>", prompt, suffix)

	return constructWithChatModel(body, content)
}

// constructWithChatModel 重写Chat请求体
func constructWithChatModel(body []byte, content string) []byte {
	// 创建新的 JSON 对象并添加到 body 中
	messages := []map[string]string{
		{
			"role":    "user",
			"content": content,
		},
	}

	body, _ = sjson.SetBytes(body, "messages", messages)

	jsonStr := string(body)
	jsonStr = strings.ReplaceAll(jsonStr, "\\u003c", "<")
	jsonStr = strings.ReplaceAll(jsonStr, "\\u003e", ">")
	return []byte(jsonStr)
}

// constructWithQwenCoderTurboModel 重写QwenCoderTurbo模型要求的请求体
func constructWithQwenCoderTurboModel(body []byte, codeMaxTokens int) []byte {
	if int(gjson.GetBytes(body, "max_tokens").Int()) > codeMaxTokens {
		body, _ = sjson.SetBytes(body, "max_tokens", codeMaxTokens)
	}

	if gjson.GetBytes(body, "n").Int() > 1 {
		body, _ = sjson.SetBytes(body, "n", 1)
	}
	prompt := gjson.GetBytes(body, "prompt")
	codeLanguage := gjson.GetBytes(body, "extra.language")

	messages := []map[string]interface{}{
		{
			"role":    "user",
			"content": "- You are a " + codeLanguage.Str + " programming expert, please complete the code appropriately based on the context, Do not generate content outside of the code." + "\n- The `#Path: ...` in the given prompt indicates the file path, which you can use to determine the programming language or simply ignore them.\n- Your maximum output tokens are: " + strconv.Itoa(codeMaxTokens) + ".",
		},
		{
			"role":    "assistant",
			"content": "```" + codeLanguage.Str + "\n" + prompt.Str,
			"partial": true,
		},
	}

	body, _ = sjson.SetBytes(body, "messages", messages)
	body, _ = sjson.DeleteBytes(body, "prompt")
	return body
}

// abortCodex 中断请求
func abortCodex(c *gin.Context, status int) {
	c.Header("Content-Type", "text/event-stream")
	c.String(status, "data: [DONE]\n")
	c.Abort()
}

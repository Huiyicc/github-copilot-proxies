package copilot

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// CodeCompletions 代码补全
func CodeCompletions(c *gin.Context) {
	ctx := c.Request.Context()

	requestID := uuid.Must(uuid.NewV4()).String()
	c.Header("x-github-request-id", requestID)

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

	c.Header("Content-Type", "text/event-stream")
	codexServiceType := os.Getenv("CODEX_SERVICE_TYPE")
	body = ConstructRequestBody(body, codexServiceType)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, os.Getenv("CODEX_API_BASE"), io.NopCloser(bytes.NewBuffer(body)))
	if nil != err {
		abortCodex(c, http.StatusInternalServerError)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	apiKeys := strings.Split(os.Getenv("CODEX_API_KEY"), ",")

	// 检查 apiKeys 是否有效
	if len(apiKeys) == 0 || (len(apiKeys) == 1 && apiKeys[0] == "") {
		abortCodex(c, http.StatusInternalServerError)
		return
	}

	randGen := rand.New(rand.NewSource(time.Now().UnixNano()))
	selectedKey := strings.TrimSpace(apiKeys[randGen.Intn(len(apiKeys))])

	if selectedKey == "" {
		abortCodex(c, http.StatusInternalServerError)
		return
	}

	req.Header.Set("Authorization", "Bearer "+selectedKey)
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
			abortCodex(c, http.StatusRequestTimeout)
			return
		}

		log.Println("request completions failed:", err.Error())
		abortCodex(c, http.StatusInternalServerError)
		return
	}
	defer CloseIO(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Println("request completions failed:", string(body))

		abortCodex(c, resp.StatusCode)
		return
	}

	c.Status(resp.StatusCode)
	// 处理 Ollama 服务的流式响应
	if codexServiceType == "ollama" {
		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					break
				}
				break
			}

			if strings.TrimSpace(line) == "" {
				continue
			}

			// json解析 line
			lineJson := gjson.Parse(line)
			uuid := uuid.Must(uuid.NewV4()).String()
			done := lineJson.Get("done").Bool()
			doneReason := lineJson.Get("done_reason").Str
			response := lineJson.Get("response").Str
			timestamp := time.Now().Unix()
			choice := map[string]interface{}{
				"text":          response,
				"index":         0,
				"logprobs":      nil,
				"finish_reason": doneReason,
			}
			choices := []map[string]interface{}{choice}
			constructLineData := map[string]interface{}{
				"id":                 uuid,
				"choices":            choices,
				"created":            timestamp,
				"model":              lineJson.Get("model").Str,
				"system_fingerprint": "fp_1c141eb703",
				"object":             "text_completion",
			}

			if done && strings.Contains(doneReason, "stop") {
				usage := map[string]interface{}{
					"prompt_tokens":            lineJson.Get("prompt_eval_count").Int(),
					"completion_tokens":        lineJson.Get("eval_count").Int(),
					"total_tokens":             lineJson.Get("prompt_eval_count").Int(),
					"prompt_cache_hit_tokens":  lineJson.Get("prompt_eval_count").Int(),
					"prompt_cache_miss_tokens": lineJson.Get("eval_count").Int(),
				}
				constructLineData["usage"] = usage
			}

			// 将修改后的数据重新编码为 JSON
			modifiedJSON, err := json.Marshal(constructLineData)
			if err != nil {
				continue
			}

			// 发送修改后的数据
			_, _ = c.Writer.WriteString("data: " + string(modifiedJSON) + "\n\n")
			c.Writer.Flush()
		}

		_, _ = c.Writer.WriteString("data: [DONE]\n\n")
		c.Writer.Flush()
		return
	}

	// 处理默认服务的响应
	_, _ = io.Copy(c.Writer, resp.Body)
}

// ConstructRequestBody 重新构建请求体
func ConstructRequestBody(body []byte, codexServiceType string) []byte {
	envCodexModel := os.Getenv("CODEX_API_MODEL_NAME")
	body, _ = sjson.SetBytes(body, "model", envCodexModel)
	body, _ = sjson.SetBytes(body, "stream", true) // 强制流式输出
	body, _ = sjson.DeleteBytes(body, "extra")
	body, _ = sjson.DeleteBytes(body, "nwo")

	// 修复ollama的coder模型不准确的问题
	if useRaw, uerr := strconv.ParseBool(os.Getenv("CODEX_USE_RAW")); uerr == nil && useRaw {
		body, _ = sjson.SetBytes(body, "raw", true)
	}
	body, _ = sjson.SetBytes(body, "options", map[string]int{
		"num_ctx": 4096,
	})
	body, _ = sjson.SetBytes(body, "keep_alive", -1)

	// 限制 prompt 和 suffix 的长度
	body = applyPromptLengthLimit(body)

	temperature, _ := strconv.ParseFloat(os.Getenv("CODEX_TEMPERATURE"), 64)
	if temperature != -1 {
		body, _ = sjson.SetBytes(body, "temperature", temperature)
	}

	codeMaxTokens, _ := strconv.Atoi(os.Getenv("CODEX_MAX_TOKENS"))
	if int(gjson.GetBytes(body, "max_tokens").Int()) > codeMaxTokens {
		body, _ = sjson.SetBytes(body, "max_tokens", codeMaxTokens)
	}

	if gjson.GetBytes(body, "n").Int() > 1 {
		body, _ = sjson.SetBytes(body, "n", 1)
	}

	// https://ollama.com/library/stable-code || https://ollama.com/library/codegemma
	if strings.Contains(envCodexModel, "stable-code") || strings.Contains(envCodexModel, "codegemma") {
		return constructWithStableCodeModel(body)
	}

	// https://ollama.com/library/codellama
	if strings.Contains(envCodexModel, "codellama") {
		return constructWithCodeLlamaModel(body)
	}

	// https://help.aliyun.com/zh/model-studio/user-guide/qwen-coder?spm=a2c4g.11186623.0.0.a5234823I6LvAG
	if strings.Contains(envCodexModel, "qwen-coder-turbo") {
		return constructWithQwenCoderTurboModel(body)
	}

	// 支持 Ollama FIM 的模型, 如:https://ollama.com/library/deepseek-coder-v2
	if codexServiceType == "ollama" {
		return constructWithOllamaModel(body, codeMaxTokens)
	}

	return body
}

// applyPromptLengthLimit 对 prompt 和 suffix 应用长度限制
func applyPromptLengthLimit(body []byte) []byte {
	envLimitPrompt := os.Getenv("CODEX_LIMIT_PROMPT")
	limitPrompt, err := strconv.Atoi(envLimitPrompt)
	if err != nil || limitPrompt <= 0 {
		return body
	}

	body = limitPromptLength(body, limitPrompt)
	body = limitSuffixLength(body, limitPrompt)

	return body
}

// limitPromptLength 限制 prompt 长度
func limitPromptLength(body []byte, limitRows int) []byte {
	prompt := gjson.GetBytes(body, "prompt")
	if !prompt.Exists() {
		return body
	}

	rows := strings.Split(prompt.Str, "\n")
	if len(rows) <= limitRows {
		return body
	}

	// 保留后面的内容
	newPrompt := strings.Join(rows[len(rows)-limitRows:], "\n")
	body, _ = sjson.SetBytes(body, "prompt", newPrompt)

	return body
}

// limitSuffixLength 限制 suffix 长度
func limitSuffixLength(body []byte, limitRows int) []byte {
	suffix := gjson.GetBytes(body, "suffix")
	if !suffix.Exists() {
		return body
	}

	rows := strings.Split(suffix.Str, "\n")
	if len(rows) <= limitRows {
		return body
	}

	// 保留前面的内容
	newSuffix := strings.Join(rows[:limitRows], "\n")
	body, _ = sjson.SetBytes(body, "suffix", newSuffix)

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
func constructWithQwenCoderTurboModel(body []byte) []byte {
	if gjson.GetBytes(body, "n").Int() > 1 {
		body, _ = sjson.SetBytes(body, "n", 1)
	}
	suffix := gjson.GetBytes(body, "suffix")
	prompt := gjson.GetBytes(body, "prompt")
	codeLanguage := gjson.GetBytes(body, "extra.language")

	messages := []map[string]interface{}{
		{
			"role":    "system",
			"content": "You are an expert in " + codeLanguage.Str + " programming, highly skilled at understanding and continuing to write code.",
		},
		{
			"role": "user",
			"content": "Combined with subsequent code snippets, help me complete the code:\n\n" +
				"Code subsequent content:\n```" + codeLanguage.Str + "\n" + suffix.Str + "```\n\n" +
				"Remember:\n" +
				"- Do not generate content outside of the code.\n" +
				"- Do not directly fill in all the code content, the maximum number of lines of code should not exceed 5 lines.\n" +
				"- Answer must refer to the code suffix content, do not exceed the boundary, otherwise repeated code will occur.\n" +
				"- If you don't know how to answer, just reply with an empty string.",
		},
		{
			"role":    "assistant",
			"content": prompt.Str,
			"partial": true,
		},
	}

	body, _ = sjson.SetBytes(body, "messages", messages)
	body, _ = sjson.DeleteBytes(body, "prompt")
	return body
}

// constructWithOllamaModel 重写Ollama模型要求的请求体
func constructWithOllamaModel(body []byte, codeMaxTokens int) []byte {
	body, _ = sjson.SetBytes(body, "options.temperature", 0)
	// stop参数处理
	stopArray := gjson.GetBytes(body, "stop").Array()
	stopSlice := make([]interface{}, len(stopArray))
	for i, v := range stopArray {
		stopSlice[i] = v.String()
	}
	body, _ = sjson.SetBytes(body, "options.stop", stopSlice)
	body, _ = sjson.SetBytes(body, "stream", true)

	maxTokens := gjson.GetBytes(body, "max_tokens").Int()
	if int(maxTokens) > codeMaxTokens {
		body, _ = sjson.SetBytes(body, "options.num_predict", codeMaxTokens)
	} else {
		body, _ = sjson.SetBytes(body, "options.num_predict", maxTokens)
	}
	return body
}

// abortCodex 中断请求
func abortCodex(c *gin.Context, status int) {
	c.Header("Content-Type", "text/event-stream")
	c.String(status, "data: [DONE]\n\n")
	c.Abort()
}

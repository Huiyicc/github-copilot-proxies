package copilot

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"ripper/internal/cache"
	"strconv"
	"time"
)

// codexCompletions 全代理GitHub的代码补全接口
func codexCompletions(c *gin.Context) {
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

	url := "https://copilot-proxy.githubusercontent.com/v1/engines/copilot-codex/completions"
	req, err := http.NewRequestWithContext(c, "POST", url, bytes.NewBuffer(body))
	if nil != err {
		abortCodex(c, http.StatusInternalServerError)
		return
	}

	// 设置Content-Type
	req.Header.Set("Content-Type", "application/json")

	token, err := getAuthToken()
	if err != nil {
		log.Println("获取GitHub Copilot的临时Token失败:", err.Error())
		abortCodex(c, http.StatusInternalServerError)
		return
	}
	req.Header.Set("authorization", "Bearer "+token)
	req.Header.Set("editor-plugin-version", "copilot-intellij/1.5.21.6667")
	req.Header.Set("copilot-language-server-version", "1.228.0")
	req.Header.Set("user-agent", "GithubCopilot/1.228.0")
	req.Header.Set("editor-version", "JetBrains-IU/242.21829.142")

	client := &http.Client{
		Timeout: 60 * time.Second,
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
		log.Println("请求GitHub官方补全接口失败:", string(body))

		abortCodex(c, resp.StatusCode)
		return
	}

	c.Status(resp.StatusCode)
	c.Header("Content-Type", "text/event-stream")
	_, _ = io.Copy(c.Writer, resp.Body)
}

// chatsCompletions 全代理GitHub的聊天补全接口
func chatsCompletions(c *gin.Context) {
	ctx := c.Request.Context()
	if ctx.Err() != nil {
		abortCodex(c, http.StatusRequestTimeout)
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if nil != err {
		abortCodex(c, http.StatusBadRequest)
		return
	}

	url := "https://api.githubcopilot.com/chat/completions"
	req, err := http.NewRequestWithContext(c, "POST", url, bytes.NewBuffer(body))
	if nil != err {
		abortCodex(c, http.StatusInternalServerError)
		return
	}

	// 固定请求头参数
	token, err := getAuthToken()
	if err != nil {
		log.Println("获取GitHub Copilot的临时Token失败:", err.Error())
		abortCodex(c, http.StatusInternalServerError)
		return
	}
	req.Header.Set("authorization", "Bearer "+token)
	req.Header.Set("editor-plugin-version", "copilot-intellij/1.5.21.6667")
	req.Header.Set("copilot-language-server-version", "1.228.0")
	req.Header.Set("user-agent", "GithubCopilot/1.228.0")
	req.Header.Set("editor-version", "JetBrains-IU/242.21829.142")
	// todo: sessionId和requestId参数暂时先不做处理, 后续有需要再添加

	client := &http.Client{
		Timeout: 60 * time.Second,
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
		log.Println("请求GitHub官方对话接口失败:", string(body))

		abortCodex(c, resp.StatusCode)
		return
	}

	c.Status(resp.StatusCode)
	c.Header("Content-Type", "text/event-stream")
	_, _ = io.Copy(c.Writer, resp.Body)
}

// getAuthToken 获取GitHub Copilot的临时Token
func getAuthToken() (string, error) {
	ghu := os.Getenv("COPILOT_GHU_TOKEN")
	cacheKey := "github:copilot_internal_v2_token:" + ghu
	token, err := cache.Get(cacheKey)
	if err != nil {
		cache.Del(cacheKey)
		return "", err
	}
	if token != nil {
		return token.(string), nil
	}

	url := "https://api.github.com/copilot_internal/v2/token"
	client := &http.Client{
		Timeout: 60 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("authorization", "token "+ghu)
	req.Header.Set("host", "api.github.com")
	req.Header.Set("accept", "*/*")
	req.Header.Set("editor-plugin-version", "copilot-intellij/1.5.21.6667")
	req.Header.Set("copilot-language-server-version", "1.228.0")
	req.Header.Set("user-agent", "GithubCopilot/1.228.0")
	req.Header.Set("editor-version", "JetBrains-IU/242.21829.142")

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("获取 Token 失败")
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	// 解析json
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", err
	}

	newToken := result["token"].(string)
	err = cache.Set(cacheKey, newToken, 1500)
	if err != nil {
		return "", err
	}
	return newToken, nil
}

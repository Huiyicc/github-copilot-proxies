package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"ripper/internal/response"
)

const (
	clientID      = "Iv1.b507a08c87ecfe98"
	deviceCodeURL = "https://github.com/login/device/code"
	tokenURL      = "https://github.com/login/oauth/access_token"
)

type githubLoginDeviceRequest struct {
	DeviceCode string `form:"device_code" json:"device_code" binding:"required"`
}

// getDeviceCode returns the device code for GitHub login.
func getDeviceCode(c *gin.Context) {
	body := map[string]string{
		"client_id": clientID,
	}

	result, err := makeRequest(c, http.MethodPost, deviceCodeURL, body)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// getGhuToken returns the GitHub user token.
func getGhuToken(c *gin.Context) {
	var params githubLoginDeviceRequest
	if err := c.ShouldBind(&params); err != nil {
		response.FailJson(c, response.FailStruct{
			Code: -1,
			Msg:  "Invalid request: " + err.Error(),
		}, false)
		return
	}

	body := map[string]string{
		"client_id":   clientID,
		"device_code": params.DeviceCode,
		"grant_type":  "urn:ietf:params:oauth:grant-type:device_code",
	}

	result, err := makeRequest(c, http.MethodPost, tokenURL, body)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// getGithubLoginDevice returns the login page for GitHub.
func getGithubLoginDevice(ctx *gin.Context) {
	ctx.Header("Content-Type", "text/html; charset=utf-8")
	ctx.HTML(http.StatusOK, "login.html", gin.H{})
}

// makeRequest makes a request to the given URL with the given method and body.
func makeRequest(c *gin.Context, method, url string, body map[string]string) (interface{}, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(c, method, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("editor-plugin-version", "copilot-intellij/1.5.21.6667")
	req.Header.Set("copilot-language-server-version", "1.228.0")
	req.Header.Set("user-agent", "GithubCopilot/1.228.0")
	req.Header.Set("editor-version", "JetBrains-IU/242.21829.142")

	httpClientTimeout, _ := time.ParseDuration(os.Getenv("HTTP_CLIENT_TIMEOUT") + "s")
	client := &http.Client{Timeout: httpClientTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

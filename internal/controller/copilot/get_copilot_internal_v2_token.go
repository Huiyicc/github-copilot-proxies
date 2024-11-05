package copilot

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"log"
	"net/http"
	"os"
	"ripper/internal/app/github_auth"
	"ripper/internal/cache"
	"strconv"
	"time"
)

// getDisguiseCopilotInternalV2Token 返回伪装的token
func getDisguiseCopilotInternalV2Token(ctx *gin.Context) {
	trackingId, _ := uuid.NewV4()
	now := time.Now().Unix()
	dcAt, _ := strconv.Atoi(os.Getenv("DISGUISE_COPILOT_TOKEN_EXPIRES_AT"))
	expiresAt := now + int64(dcAt)
	sku := "copilot_for_business_seat"

	copilotToken := github_auth.JsonMap2SignToken(map[string]interface{}{
		"tid":  trackingId,
		"exp":  expiresAt,
		"sku":  sku,
		"st":   "dotcom",
		"chat": 1,
		"u":    "github",
	})

	endpoints := make(map[string]interface{})
	endpoints["api"] = os.Getenv("API_BASE_URL")
	endpoints["origin-tracker"] = "https://origin-tracker.githubusercontent.com"
	endpoints["proxy"] = os.Getenv("PROXY_BASE_URL")
	endpoints["telemetry"] = os.Getenv("TELEMETRY_BASE_URL")

	gout := gin.H{
		"annotations_enabled":                      true,
		"chat_enabled":                             true,
		"chat_jetbrains_enabled":                   true,
		"code_quote_enabled":                       true,
		"code_review_enabled":                      true,
		"codesearch":                               true,
		"copilot_ide_agent_chat_gpt4_small_prompt": false,
		"copilotignore_enabled":                    false,
		"endpoints":                                endpoints,
		"expires_at":                               expiresAt,
		"individual":                               true,
		"nes_enabled":                              false,
		"prompt_8k":                                true,
		"public_suggestions":                       "disabled",
		"refresh_in":                               1500,
		"sku":                                      sku,
		"snippy_load_test_enabled":                 false,
		"telemetry":                                "disabled",
		"token":                                    copilotToken,
		"tracking_id":                              trackingId,
		"intellij_editor_fetcher":                  false,
		"vsc_electron_fetcher":                     false,
		"vs_editor_fetcher":                        false,
		"vsc_panel_v2":                             false,
		"xcode":                                    true,
		"xcode_chat":                               true,
	}
	ctx.JSON(http.StatusOK, gout)
}

// getCopilotInternalV2Token 获取github copilot官方token
func getCopilotInternalV2Token(c *gin.Context) {
	ghu := os.Getenv("COPILOT_GHU_TOKEN")
	if ghu == "" {
		log.Println("ghu token is empty")
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"message": "ghu token is empty",
		})
		return
	}

	cacheKey := "copilot_internal_v2_token"
	token, err := cache.Get(cacheKey)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		cache.Del(cacheKey)
		return
	}

	if token != nil {
		c.JSON(http.StatusOK, token)
		return
	}

	url := "https://api.github.com/copilot_internal/v2/token"
	req, err := http.NewRequestWithContext(c, "GET", url, nil)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	req.Header.Set("authorization", "token "+ghu)
	req.Header.Set("editor-plugin-version", "copilot-intellij/1.5.21.6667")
	req.Header.Set("editor-version", "JetBrains-IU/242.21829.142")
	req.Header.Set("user-agent", "GithubCopilot/1.228.0")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err.Error())
		c.JSON(resp.StatusCode, gin.H{"error": err.Error()})
		return
	}
	if resp.StatusCode != 200 {
		errorMsg := "获取 Token 失败, 当前 ghu_token 账户可能并未订阅 github copilot 服务!"
		c.JSON(resp.StatusCode, gin.H{"error": errorMsg})
		log.Println(errorMsg)
		return
	}
	defer resp.Body.Close()

	var result interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	cache.Set(cacheKey, result, 1500)
	c.JSON(resp.StatusCode, result)
}

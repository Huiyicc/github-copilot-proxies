package copilot

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"log"
	"math/rand"
	"net/http"
	"os"
	"ripper/internal/app/github_auth"
	"ripper/internal/cache"
	"strconv"
	"strings"
	"time"
)

// GetDisguiseCopilotInternalV2Token 返回伪装的token
func GetDisguiseCopilotInternalV2Token(ctx *gin.Context) {
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
	endpoints["origin-tracker"] = "https://origin-tracker.individual.githubcopilot.com"
	endpoints["proxy"] = os.Getenv("PROXY_BASE_URL")
	endpoints["telemetry"] = os.Getenv("TELEMETRY_BASE_URL")

	gout := gin.H{
		"annotations_enabled":                      true,
		"chat_enabled":                             true,
		"chat_jetbrains_enabled":                   true,
		"code_quote_enabled":                       true,
		"code_review_enabled":                      false,
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
		"limited_user_quotas":                      nil,
		"limited_user_reset_date":                  nil,
		"vsc_electron_fetcher_v2":                  false,
	}
	ctx.JSON(http.StatusOK, gout)
}

// GetCopilotInternalV2Token 获取github copilot官方token
func GetCopilotInternalV2Token(c *gin.Context) {
	ghuTokens := strings.Split(os.Getenv("COPILOT_GHU_TOKEN"), ",")
	if len(ghuTokens) == 0 {
		return
	}

	rand.Seed(time.Now().UnixNano())
	ghu := ghuTokens[rand.Intn(len(ghuTokens))]
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

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err.Error())
		c.JSON(resp.StatusCode, gin.H{"error": err.Error()})
		return
	}
	if resp.StatusCode != 200 {
		errorMsg := "获取 Token 失败, 当前 ghu_token 账户可能并未订阅 github copilot 服务!" + ghu
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

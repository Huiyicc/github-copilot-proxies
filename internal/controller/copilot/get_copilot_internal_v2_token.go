package copilot

import (
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"net/http"
	"os"
	"ripper/internal/app/github_auth"
	"time"
)

func getCopilotInternalV2Token(ctx *gin.Context) {
	trackingId, _ := uuid.NewV4()
	now := time.Now().Unix()
	expiresAt := now + 1800
	sku := "copilot_for_business_seat"

	copilotToken := github_auth.JsonMap2SignToken(map[string]interface{}{
		"tid":  trackingId,
		"exp":  expiresAt,
		"sku":  sku,
		"st":   "dotcom",
		"chat": 1,
		"u":    "fakeuser",
	})

	endpoints := make(map[string]interface{})
	endpoints["api"] = os.Getenv("API_BASE_URL")
	endpoints["origin-tracker"] = "https://origin-tracker.githubusercontent.com"
	endpoints["proxy"] = os.Getenv("PROXY_BASE_URL")
	endpoints["telemetry"] = os.Getenv("TELEMETRY_BASE_URL")

	gout := gin.H{
		"annotations_enabled":                      false,
		"chat_enabled":                             true,
		"chat_jetbrains_enabled":                   true,
		"code_quote_enabled":                       true,
		"codesearch":                               false,
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
		"project":                                  "copilot-proxy",
	}
	ctx.JSON(http.StatusOK, gout)
}

package auth

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"ripper/internal/app/github_auth"
	"ripper/internal/cache"
	"ripper/internal/middleware"
	"ripper/internal/response"
	jwtpkg "ripper/pkg/jwt"
	"time"
)

type getLoginOauthAuthorizeRequest struct {
	ClientId    string `json:"client_id" form:"client_id"`
	Prompt      string `json:"prompt" form:"prompt"`
	RedirectUri string `json:"redirect_uri" form:"redirect_uri"`
	Scope       string `json:"scope" form:"scope"`
	State       string `json:"state" form:"state"`
}

func getLoginOauthAuthorize(ctx *gin.Context) {
	req := getLoginOauthAuthorizeRequest{}
	err := ctx.BindQuery(&req)
	if err != nil {
		response.FailJson(ctx, response.FailStruct{
			Code: -1,
			Msg:  "Invalid request.",
		}, false)
		return
	}

	vsCopilotClientId := os.Getenv("VS_COPILOT_CLIENT_ID")
	if req.ClientId != vsCopilotClientId {
		response.FailJson(ctx, response.FailStruct{
			Code: -1,
			Msg:  "Invalid client id.",
		}, false)
		return
	}

	oauthCode := github_auth.GenDevicesCode(20)
	cai := github_auth.ClientOAuthInfo{
		ClientId: req.ClientId,
		Code:     oauthCode,
		Scope:    req.Scope,
	}
	cacheKey := "oauth2_authorize_" + req.ClientId
	caiInfo, _ := json.Marshal(cai)
	err = cache.Set(cacheKey, caiInfo, 300)
	if err != nil {
		response.FailJson(ctx, response.FailStruct{
			Code: -1,
			Msg:  "Internal error.",
		}, false)
		return
	}

	// Redirect to the client's redirect_uri
	browserSessionId := github_auth.GenDevicesCode(64)
	ctx.Redirect(302, req.RedirectUri+"?browserSessionId="+browserSessionId+"&code="+oauthCode+"&state="+req.State)
}

func postLoginOauthAccessTokenForVs2022(ctx *gin.Context) {
	v, exists := ctx.Get("client_auth_info")
	if !exists {
		response.FailJson(ctx, response.FailStruct{
			Code: -1,
			Msg:  "Invalid client id.",
		}, false)
		return
	}
	cliAuthInfo := v.(*github_auth.ClientOAuthInfo)
	t := time.Now()
	t.Add(24 * 3 * time.Hour)
	tk, _ := jwtpkg.CreateToken(&middleware.UserLoad{
		CardCode:         cliAuthInfo.Code,
		Client:           cliAuthInfo.ClientId,
		RegisteredClaims: jwtpkg.CreateStandardClaims(t.Unix(), "user"),
	})
	ctx.JSON(http.StatusOK, gin.H{
		"access_token": tk,
		"scope":        cliAuthInfo.Scope,
		"token_type":   "bearer",
	})
}

func getSiteSha(ctx *gin.Context) {
	ctx.Header("X-GitHub-Request-Id", "C0E1:6A1A:1A1F:2A1D:1A1F:1A1F:1A1F:1A1F")
	ctx.JSON(http.StatusOK, gin.H{})
}

func getLoginConfig(ctx *gin.Context) {
	loginPassword := os.Getenv("LOGIN_PASSWORD")
	ctx.JSON(http.StatusOK, gin.H{
		"is_login_password": loginPassword != "",
	})
}

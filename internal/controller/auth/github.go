package auth

import (
	_ "embed"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"ripper/internal/app/github_auth"
	"ripper/internal/middleware"
	"ripper/internal/response"
	jwtpkg "ripper/pkg/jwt"
	"time"
)

type postLoginDeviceCodeRequest struct {
	ClientId string `json:"client_id" form:"client_id"`
}

type postLoginDeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`      // 设备代码
	UserCode        string `json:"user_code"`        // 用户代码
	VerificationUrl string `json:"verification_uri"` // 验证地址
	ExpiresIn       int    `json:"expires_in"`       // 过期时间
	Interval        int    `json:"interval"`         // 间隔时间
}

func postLoginDeviceCode(ctx *gin.Context) {
	cli := postLoginDeviceCodeRequest{}
	if err := ctx.ShouldBind(&cli); err != nil {
		response.FailJson(ctx, response.FailStruct{
			Code: -1,
			Msg:  "Invalid client id.",
		}, false)
		return
	}

	if cli.ClientId == "" {
		response.FailJson(ctx, response.FailStruct{
			Code: -1,
			Msg:  "Client id is required.",
		}, false)
		return
	}

	uid, devid, err := github_auth.BindClientToCode(cli.ClientId, 1800)
	if err != nil {
		response.FailJson(ctx, response.FailStruct{
			Code: -1,
			Msg:  err.Error(),
		}, false)
		return
	}
	ctx.JSON(http.StatusOK, postLoginDeviceCodeResponse{
		DeviceCode:      devid,
		UserCode:        uid,
		VerificationUrl: fmt.Sprintf("%s/login/device?user_code=%s", os.Getenv("DEFAULT_BASE_URL"), uid),
		ExpiresIn:       1800,
		Interval:        5,
	})
}

func postLoginOauthAccessToken(ctx *gin.Context) {
	v, exists := ctx.Get("client_auth_info")
	if !exists {
		ctx.JSON(http.StatusOK, gin.H{
			"error":             "authorization_pending",
			"error_description": "The authorization request is still pending.",
			"error_uri":         "https://docs.github.com/developers/apps/authorizing-oauth-apps#error-codes-for-the-device-flow",
		})
		return
	}
	cliAuthInfo := v.(*github_auth.ClientAuthInfo)
	t := time.Now()
	t.Add(24 * 3 * time.Hour)
	u, err := github_auth.GetClientAuthInfo(cliAuthInfo.UserCode)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"error":             "access_denied",
			"error_description": "You must make a new request for a device code.",
			"error_uri":         "https://docs.github.com/developers/apps/authorizing-oauth-apps#error-codes-for-the-device-flow",
		})
		return
	}
	tk, _ := jwtpkg.CreateToken(&middleware.UserLoad{
		CardCode:         u.CardCode,
		Client:           cliAuthInfo.ClientId,
		RegisteredClaims: jwtpkg.CreateStandardClaims(t.Unix(), "user"),
	})
	_ = github_auth.RemoveClientAuthInfoByDeviceCode(cliAuthInfo.ClientId)
	ctx.JSON(http.StatusOK, gin.H{
		"access_token": tk,
		"scope":        "",
		"token_type":   "bearer",
	})
}

type loginDeviceRequestInfo struct {
	Code          string `json:"code"`
	Authorization string `json:"authorization"`
}

func postLoginDevice(ctx *gin.Context) {
	var info loginDeviceRequestInfo
	if err := response.BindStruct(ctx, &info); err != nil {
		response.FailJson(ctx, response.FailStruct{
			Code: 422,
			Msg:  "Invalid json.",
		}, false)
		return
	}
	// 检查code是否存在
	authInfo, err := github_auth.GetClientAuthInfo(info.Code)
	if err != nil {
		response.FailJson(ctx, response.FailStruct{
			Code: 422,
			Msg:  "Invalid code.",
		}, false)
		return
	}

	err = github_auth.UpdateClientAuthStatusByDeviceCode(authInfo.DeviceCode, info.Authorization)
	if err != nil {
		response.FailJson(ctx, response.FailStruct{
			Code: 500,
			Msg:  "System Error",
		}, false)
		return
	}
	response.SuccessJson(ctx, "ok")
}

func getLoginDevice(ctx *gin.Context) {
	ctx.Header("Content-Type", "text/html; charset=utf-8")
	ctx.HTML(http.StatusOK, "code.html", gin.H{})
}

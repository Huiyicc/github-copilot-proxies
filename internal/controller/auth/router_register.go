package auth

import (
	"github.com/gin-gonic/gin"
	"ripper/internal/middleware"
	"strings"
)

func GinApi(g *gin.RouterGroup) {
	// 启动设备代码登录流程
	g.POST("/login/device/code", postLoginDeviceCode)
	g.POST("/login/device", postLoginDevice)
	g.GET("/login/device", getLoginDevice)
	g.POST("/login/oauth/access_token", func(ctx *gin.Context) {
		if strings.Index(ctx.Request.UserAgent(), "VSTeamExplorer") != -1 {
			middleware.AuthCodeFlowCheckAuth(ctx)
		} else {
			middleware.DeviceCodeCheckAuth(ctx)
		}
	}, func(ctx *gin.Context) {
		if strings.Index(ctx.Request.UserAgent(), "VSTeamExplorer") != -1 {
			postLoginOauthAccessTokenForVs2022(ctx)
		} else {
			postLoginOauthAccessToken(ctx)
		}
	})

	// oauth2 登录
	g.GET("/login/oauth/authorize", getLoginOauthAuthorize)

	// enterprise 验证
	g.GET("/site/sha", getSiteSha)
}

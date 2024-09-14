package auth

import (
	"github.com/gin-gonic/gin"
	"ripper/internal/middleware"
)

func GinApi(g *gin.RouterGroup) {
	// 启动设备代码登录流程
	g.POST("/login/device/code", postLoginDeviceCode)
	g.POST("/login/device", postLoginDevice)
	g.GET("/login/device", getLoginDevice)
	g.POST("/login/oauth/access_token", middleware.DeviceCodeCheckAuth, postLoginOauthAccessToken)
}

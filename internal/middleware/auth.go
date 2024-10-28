package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"ripper/internal/app/github_auth"
	"ripper/internal/response"
	jwtpkg "ripper/pkg/jwt"
	"strconv"
	"strings"
	"time"
)

type OAuthCheck struct {
	ClientId   string `json:"client_id" form:"client_id"`
	DeviceCode string `json:"device_code" form:"device_code"`
	GrantType  string `json:"grant_type" form:"grant_type"`
}

func DeviceCodeCheckAuth(ctx *gin.Context) {
	checkInfo := &OAuthCheck{}
	if err := ctx.ShouldBind(&checkInfo); err != nil {
		response.FailJson(ctx, response.FailStruct{
			Code: -1,
			Msg:  "Invalid client id.",
		}, false)
		ctx.Abort()
		return
	}
	info, _ := github_auth.GetClientAuthInfoByDeviceCode(checkInfo.DeviceCode)
	if info.CardCode == "" {
		ctx.JSON(http.StatusOK, gin.H{
			"error":             "authorization_pending",
			"error_description": "The authorization request is still pending.",
			"error_uri":         "https://docs.github.com/developers/apps/authorizing-oauth-apps#error-codes-for-the-device-flow",
		})
		ctx.Abort()
		return
	}
	ctx.Set("client_auth_info", info)
	ctx.Next()
}

func AuthCodeFlowCheckAuth(ctx *gin.Context) {
	checkInfoClient := &github_auth.ClientOAuthInfo{}
	err := ctx.Bind(&checkInfoClient)
	if err != nil {
		response.FailJson(ctx, response.FailStruct{
			Code: -1,
			Msg:  "Invalid client id.",
		}, false)
		ctx.Abort()
		return
	}
	oauthCodeInfo, err := github_auth.GetOAuthCodeInfoByClientIdAndCode(checkInfoClient.ClientId, checkInfoClient.Code)
	if err != nil {
		response.FailJson(ctx, response.FailStruct{
			Code: -1,
			Msg:  "Invalid client id.",
		}, false)
		ctx.Abort()
		return
	}

	ctx.Set("client_auth_info", oauthCodeInfo)
	ctx.Next()
}

func AccessTokenCheckAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("Authorization")
		if token == "" {
			response.FailJsonAndStatusCode(c, http.StatusForbidden, response.NoAccess, false)
			c.Abort()
			return
		}
		last := strings.Index(token, " ")
		if len(token) < last || last == -1 {
			response.FailJsonAndStatusCode(c, http.StatusForbidden, response.TokenWrongful, false)
			c.Abort()
			return
		}
		token = token[last+1:]
		chk, jwter, err := jwtpkg.CheckToken(token, &UserLoad{}, "user")
		if err != nil {
			errmsg := response.TokenWrongful
			errmsg.Msg = "令牌验证错误"
			response.FailJsonAndStatusCode(c, http.StatusForbidden, errmsg, true, err.Error())
			c.Abort()
			return
		}
		if !chk {
			response.FailJsonAndStatusCode(c, http.StatusForbidden, response.NoAccess, true, "破损令牌")
			c.Abort()
			return
		}
		chs := true
		issuerStr := ""
		issuerStr, err = jwter.GetIssuer()
		if err != nil {
			chs = false
			c.Abort()
			return
		}
		if "user" != issuerStr && issuerStr != "" {
			chs = false
			c.Abort()
			return
		}
		if !chs {
			errmsg := response.TokenWrongful
			errmsg.Msg = "签名错误"
			response.FailJsonAndStatusCode(c, http.StatusForbidden, errmsg, true, err.Error())
			c.Abort()
			return
		}
		c.Set("token", jwter)
		c.Set("tokenStr", token)
		c.Set("token.issuer", issuerStr)
		c.Next()
	}
}

func TokenCheckAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientType := os.Getenv("COPILOT_CLIENT_TYPE")
		copilotProxyAll, err := strconv.ParseBool(os.Getenv("COPILOT_PROXY_ALL"))
		if clientType == "github" && !copilotProxyAll {
			c.Next()
			return
		}
		token := c.Request.Header.Get("Authorization")
		if token == "" {
			response.FailJsonAndStatusCode(c, http.StatusUnauthorized, response.TokenWrongful, false)
			c.Abort()
			return
		}
		last := strings.Index(token, " ")
		if len(token) < last || last == -1 {
			response.FailJsonAndStatusCode(c, http.StatusUnauthorized, response.TokenWrongful, false)
			c.Abort()
			return
		}
		token = token[last+1:]
		parsedToken := parseAuthorizationToken(token)
		// 校验exp是否过期
		expired, err := isExpired(parsedToken["exp"])
		if err != nil {
			response.FailJsonAndStatusCode(c, http.StatusUnauthorized, response.TokenWrongful, false)
			c.Abort()
			return
		} else {
			if expired {
				response.FailJsonAndStatusCode(c, http.StatusUnauthorized, response.TokenOverdue, false)
				c.Abort()
				return
			}
		}
		rawToken := github_auth.JsonMap2Token(map[string]interface{}{
			"tid":  parsedToken["tid"],
			"exp":  parsedToken["exp"],
			"sku":  parsedToken["sku"],
			"st":   parsedToken["st"],
			"chat": parsedToken["chat"],
			"u":    parsedToken["u"],
		})
		sign := "1:" + github_auth.Token2Sign(rawToken)
		if sign != parsedToken["8kp"] {
			response.FailJsonAndStatusCode(c, http.StatusUnauthorized, response.TokenWrongful, false)
			c.Abort()
			return
		}
		c.Next()
	}
}

func parseAuthorizationToken(token string) map[string]string {
	result := make(map[string]string)
	pairs := strings.Split(token, ";")

	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			key := kv[0]
			value := kv[1]

			if key == "tid" || key == "exp" || key == "sku" || key == "st" || key == "8kp" || key == "chat" || key == "u" {
				result[key] = value
			}
		}
	}

	return result
}

func isExpired(expStr string) (bool, error) {
	exp, err := strconv.ParseInt(expStr, 10, 64)
	if err != nil {
		return false, fmt.Errorf("invalid exp timestamp: %v", err)
	}

	now := time.Now().Unix()
	return now > exp, nil
}

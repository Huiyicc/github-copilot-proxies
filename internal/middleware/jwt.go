package middleware

import (
	"github.com/gin-gonic/gin"
	"hzer/internal/response"
	jwtpkg "hzer/pkg/jwt"
)

// JWTCheck 检查是否登陆
// 检查完毕会将jwt结构体写入到Context
// 适用于同时用于公开与鉴权的路由
func JWTCheck(c *gin.Context, model jwtpkg.LoadModel, issure ...string) (bool, error) {
	token := c.Request.Header.Get("Authorization")
	if token == "" {
		return false, nil
	}
	if len(token) < 8 {
		return false, nil
	}
	token = token[7:]
	chk, jwter, err := jwtpkg.CheckToken(token, model, "")
	if err != nil {
		return false, err
	}
	chs := true
	for _, v := range issure {
		jwt, err := jwter.GetIssuer()
		if err != nil {
			chs = false
			break
		}
		if v != jwt {
			chs = false
			break
		}
	}
	if !chs {
		return false, nil
	}
	if !chk {
		return false, nil
	}
	c.Set("token", jwter)
	c.Next()
	return true, nil
}

// JWTAuth 为JWT中间件，客户端下需要在header带上Authorization: Bearer <token>
// issure 为可选验证签名，支持多参选择
func JWTAuth(model jwtpkg.LoadModel, issure ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("Authorization")
		if token == "" {
			response.FailJson(c, response.NoAccess, false)
			c.Abort()
			return
		}
		if len(token) < 8 {
			response.FailJson(c, response.TokenWrongful, false)
			c.Abort()
			return
		}
		token = token[7:]
		chk, jwter, err := jwtpkg.CheckToken(token, model, "")
		if err != nil {
			errmsg := response.TokenWrongful
			errmsg.Msg = "令牌验证错误"
			response.FailJson(c, errmsg, true, err.Error())
			c.Abort()
			return
		}
		if !chk {
			response.FailJson(c, response.NoAccess, true, "破损令牌")
			c.Abort()
			return
		}
		chs := true
		issuerStr := ""
		for _, v := range issure {
			issuerStr, err = jwter.GetIssuer()
			if err != nil {
				chs = false
				break
			}
			if v != issuerStr {
				chs = false
				break
			}
		}
		if !chs {
			errmsg := response.TokenWrongful
			errmsg.Msg = "签名错误"
			response.FailJson(c, errmsg, true, err.Error())
			c.Abort()
			return
		}
		c.Set("token", jwter)
		c.Set("tokenStr", token)
		c.Set("token.issuer", issuerStr)
		c.Next()
	}
}

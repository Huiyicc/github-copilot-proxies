package jwt_test

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"hzer/internal/middleware"
	"hzer/internal/response"
	jwtpkg "hzer/pkg/jwt"
)

// 业务模型
type UserLoad struct {
	ID uint `json:"id"`
	jwt.RegisteredClaims
}

// 快速创建接口
func newUserLoad(ID uint, ExpiresAt int64, Issuer string) *UserLoad {
	return &UserLoad{
		ID:               ID,
		RegisteredClaims: jwtpkg.CreateStandardClaims(ExpiresAt, Issuer),
	}
}

func testHandler(c *gin.Context) {
	// 获取jwt数据并处理错误
	// 可以在中间件完成这一步,根据业务自行扩展
	token, err := jwtpkg.GetJwtProto(c, &UserLoad{})
	if err != nil {
		response.FailJson(c, response.FailStruct{
			Code: 401,
			Msg:  err.Error(),
		}, false)
	}
	fmt.Println(token)
}

func main() {
	r := gin.Default()

	//使用中间件时绑定业务jwt模型,并且可以设置验证的签发人
	r.Use(middleware.JWTAuth(&UserLoad{}, "is"))
	r.GET("/test", testHandler)
	r.Run()
}

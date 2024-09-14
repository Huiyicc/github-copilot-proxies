package jwt

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"reflect"
	"time"
)

var SecretKey string

// LoadModel 成员中套上jwt.RegisteredClaims
type LoadModel interface {
	GetExpirationTime() (*jwt.NumericDate, error)
	GetIssuedAt() (*jwt.NumericDate, error)
	GetNotBefore() (*jwt.NumericDate, error)
	GetIssuer() (string, error)
	GetSubject() (string, error)
	GetAudience() (jwt.ClaimStrings, error)
}

// JWT jwt对象
type JWT struct {
	// 声明签名信息
	SigningKey []byte
}

// NewJWT 初始化jwt对象
func NewJWT() *JWT {
	return &JWT{
		[]byte(SecretKey),
	}
}

// CreateToken 调用jwt-go库生成token,编码的算法为jwt.SigningMethodHS256
func (j *JWT) CreateToken(payload jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	return token.SignedString(j.SigningKey)
}

// ParserToken 将token解码并验证
func (j *JWT) ParserToken(tokenString string, model LoadModel) (jwt.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, model, func(token *jwt.Token) (interface{}, error) {
		return j.SigningKey, nil
	})
	if err != nil {
		return nil, err
	}
	_, err = parserToken(token, model)
	if token.Valid {
		return token.Claims, nil
	}
	return nil, fmt.Errorf("token无效")
}

func parserToken[T LoadModel](token *jwt.Token, model T) (jwt.Claims, error) {
	if claims, ok := token.Claims.(T); !ok {
		return nil, fmt.Errorf("token结构错误")
	} else {
		return claims, nil
	}
}

// CreateToken 用于快速生成一个Token
// UserLoad 用户负载
// ExpiresAt 多少秒之后过期，单位：秒
// Issuer 签名颁发着
func CreateToken(JWTLoad jwt.Claims) (Token string, err error) {
	// 构造SignKey: 签名和解签名需要使用一个值
	j := NewJWT()
	// 构造用户claims信息(负荷)
	token, err := j.CreateToken(JWTLoad)
	if err != nil {
		return "", err
	}
	return token, nil
}

// CreateStandardClaims 快速创建签名数据
func CreateStandardClaims(ExpiresAt int64, Issuer string) jwt.RegisteredClaims {
	return jwt.RegisteredClaims{
		Issuer:    Issuer,                                                                     // 签名颁发者
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * time.Duration(ExpiresAt))), // 签名过期时间
		IssuedAt:  jwt.NewNumericDate(time.Now()),                                             //
		NotBefore: jwt.NewNumericDate(time.Now()),                                             // 签名生效时间
	}
}

// CheckToken 用于检查Token是否有效
// issuer为可选参数
func CheckToken[T LoadModel](token string, model T, issuer string) (bool, T, error) {
	j := NewJWT()
	load, err := j.ParserToken(token, model)
	if err != nil {
		return false, model, err
	}
	expTime, err := load.GetExpirationTime()
	if expTime.Unix() < time.Now().Unix() {
		return false, model, errors.New("token已过期，请重新登录")
	}
	loadIssuer, err := load.GetIssuer()
	if err != nil {
		return false, model, err
	}
	if issuer == loadIssuer {
		return true, load.(T), nil
	}
	return false, model, errors.New("token签名错误，请重新登录")
}

// GetJwtProto 从Gin中获取Jwt原型体
// model为业务内jwt模型
func GetJwtProto[T any](c *gin.Context, model T) (T, error) {
	tokener, _ := c.Get("token")
	if tokener == nil {
		return model, errors.New("JWT Is NULL")
	}
	token, ok := tokener.(T)
	if !ok {
		return model, errors.New("JWTLoad Error")
	}
	return token, nil
}

// GetTokenLoad 用于从GinContext中取回JWTUsermapm
func GetTokenLoad(c *gin.Context) (*JWTLoad, map[string]interface{}) {
	token, _ := c.Get("token")
	if token == nil {
		return nil, nil
	}
	load := token.(*JWTLoad)
	if load.UserLoad == nil {
		return nil, nil
	}
	return load, load.UserLoad.(map[string]interface{})
}

func ShouldBindTokenLoad(c *gin.Context, obj any) error {
	var e error
	token, _ := c.Get("token")
	if token == nil {
		e = errors.New("token is invalid")
	}
	load := token.(*JWTLoad)
	if load.UserLoad == nil {
		e = errors.New("token is illegal")
	}
	for k, v := range load.UserLoad.(map[string]interface{}) {
		fmt.Println(v, k)
		e = SetField(obj, k, v)
		if e != nil {
			return e
		}
	}
	return nil
}
func SetField(obj interface{}, name string, value interface{}) error {
	structValue := reflect.ValueOf(obj).Elem()
	structFieldValue := structValue.FieldByName(name)
	fmt.Println(structFieldValue)

	if !structFieldValue.IsValid() {
		return fmt.Errorf("no such field: %s in obj", name)
	}

	if !structFieldValue.CanSet() {
		return fmt.Errorf("cannot set %s field value", name)
	}

	structFieldType := structFieldValue.Type()
	val := reflect.ValueOf(value)
	if structFieldType != val.Type() {
		return errors.New("provided value type didn't match obj field type")
	}

	structFieldValue.Set(val)
	return nil
}

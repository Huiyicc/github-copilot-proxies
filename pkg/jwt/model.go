package jwt

import "github.com/golang-jwt/jwt/v5"

type JWTLoad struct {
	UserLoad interface{} `json:"user_load"`
	Version  int64
	jwt.RegisteredClaims
}

package middleware

import (
	"github.com/golang-jwt/jwt/v5"
	jwtPkg "hzer/pkg/jwt"
)

type AdminLoad struct {
	Username string `json:"username"`
}

type UserLoad struct {
	CardCode string `json:"token"`
	Client   string `json:"client"`
	jwt.RegisteredClaims
}

func NewUserLoad(ID uint, ExpiresAt int64, Issuer string) *UserLoad {
	return &UserLoad{
		RegisteredClaims: jwtPkg.CreateStandardClaims(ExpiresAt, Issuer),
	}
}

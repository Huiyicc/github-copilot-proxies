package middleware

import (
	"github.com/golang-jwt/jwt/v5"
	jwtPkg "ripper/pkg/jwt"
)

type AdminLoad struct {
	Username string `json:"username"`
}

type UserLoad struct {
	UserDisplayName string `json:"userDisplayName,omitempty"`
	CardCode        string `json:"token"`
	Client          string `json:"client"`
	jwt.RegisteredClaims
}

func NewUserLoad(ID uint, ExpiresAt int64, Issuer string) *UserLoad {
	return &UserLoad{
		RegisteredClaims: jwtPkg.CreateStandardClaims(ExpiresAt, Issuer),
	}
}

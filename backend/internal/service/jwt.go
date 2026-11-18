package service

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID string `json:"user_id"` // 文档型 _id，如 u_xxx
	OpenID string `json:"openid"`
	jwt.RegisteredClaims
}

func generateTokenInternal(userID, openid, secret string) (string, error) {
	claims := Claims{
		UserID: userID,
		OpenID: openid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(168 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

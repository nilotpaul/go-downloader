package types

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

type Session struct {
	Token     jwt.Token `json:"token"`
	UserID    string    `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

type JWTSession struct {
	UserID    string    `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

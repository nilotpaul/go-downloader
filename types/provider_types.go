package types

import (
	"github.com/gofiber/fiber/v2"
	"golang.org/x/oauth2"
)

type OAuthProvider interface {
	Authenticate(authCode string) error
	GetAccessToken() string
	GetRefreshToken() string
	RefreshToken(c *fiber.Ctx, userID string) (*oauth2.Token, error)
	IsTokenValid() bool
	GetAuthURL(state string) string
	CreateOrUpdateAccount() (string, error)
	CreateSession(c *fiber.Ctx, userID string) error
	UpdateTokens(acc *GoogleAccount) error
}

type ProviderRegistry interface {
	Register(providerName string, p OAuthProvider)
	GetProvider(providerName string) (OAuthProvider, error)
}

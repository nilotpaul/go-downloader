package types

import (
	"github.com/gofiber/fiber/v2"
	"golang.org/x/oauth2"
)

type OAuthProvider interface {
	Authenticate(string) error
	GetAccessToken() string
	GetRefreshToken() string
	RefreshToken(*fiber.Ctx, string, bool) (*oauth2.Token, error)
	IsTokenValid() bool
	GetAuthURL(state string) string
	CreateOrUpdateAccount() (string, error)
	CreateSession(c *fiber.Ctx, userID string) error
	UpdateTokens(*GoogleAccount) error
}

type ProviderRegistry interface {
	Register(string, OAuthProvider)
	GetProvider(string) (OAuthProvider, error)
}

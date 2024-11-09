package store

import (
	"database/sql"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/nilotpaul/go-downloader/config"
	"github.com/nilotpaul/go-downloader/types"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func TestNewGoogleProvider(t *testing.T) {
	cfg := googleProviderConfig{
		googleClientID:     "mock_client_id",
		googleClientSecret: "mock_client_secret",
		googleRedirectURL:  "http://localhost:3000/api/v1/callback/google",
	}

	// Mock env vars
	env := config.EnvConfig{}

	// Mock the database connection
	db := sql.DB{}

	provider := NewGoogleProvider(cfg, &db, env)
	assert.NotNil(t, provider)
	assert.Equal(t, cfg.googleClientID, provider.Config.ClientID)
	assert.Equal(t, cfg.googleClientSecret, provider.Config.ClientSecret)
	assert.Equal(t, cfg.googleClientID, provider.Config.ClientID)
	assert.Equal(t, cfg.googleRedirectURL, provider.Config.RedirectURL)
	assert.Equal(t, scopes, provider.Config.Scopes)
}

func TestNewGoogleProvider_GetAccessToken(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	// Mock GoogleProvider
	gp := &GoogleProvider{}

	a.Equal("", gp.GetAccessToken())

	// Mock OAuth Token
	mockToken := &oauth2.Token{
		AccessToken:  "mock_access_token",
		RefreshToken: "mock_refresh_token",
		Expiry:       time.Now().Add(time.Hour), // Mock expiry of one hour
	}
	gp.Token = mockToken
	a.Equal(mockToken.AccessToken, gp.GetAccessToken())
}

func TestNewGoogleProvider_GetRefreshToken(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	// Mock GoogleProvider
	gp := &GoogleProvider{}

	a.Equal("", gp.GetRefreshToken())

	// Mock OAuth Token
	mockToken := &oauth2.Token{
		AccessToken:  "mock_access_token",
		RefreshToken: "mock_refresh_token",
		Expiry:       time.Now().Add(time.Hour), // Mock expiry of one hour
	}
	gp.Token = mockToken
	a.Equal(mockToken.RefreshToken, gp.GetRefreshToken())
}

func TestNewGoogleProvider_IsTokenValid(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	// Mock GoogleProvider
	gp := &GoogleProvider{}

	a.Nil(gp.Token)
	a.False(gp.IsTokenValid())

	// Mock OAuth Token
	mockToken := &oauth2.Token{
		AccessToken:  "mock_access_token",
		RefreshToken: "mock_refresh_token",
		Expiry:       time.Now().Add(time.Hour), // Mock expiry of one hour
	}
	gp.Token = mockToken

	a.True(gp.IsTokenValid())
}

func TestNewGoogleProvider_GetAuthURL(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	// Mock OAuth Config
	mockConfig := &oauth2.Config{
		ClientID:     "mock_client_id",
		ClientSecret: "mock_client_secret",
		Endpoint:     google.Endpoint,
		RedirectURL:  "http://localhost:3000/api/v1/callback/google",
		Scopes:       scopes,
	}
	// Mock GoogleProvider
	gp := &GoogleProvider{Config: mockConfig}

	// Mock Random State
	mockState := "mock_state"
	authURL := gp.GetAuthURL(mockState)

	// authURL is URL-Encoded, converting it in plain non-encoded URL string.
	decodedURL, err := url.QueryUnescape(authURL)
	a.Nil(err)

	// Example:
	// https://accounts.google.com/o/oauth2/auth?access_type=offline&client_id=mock_client_id&prompt=consent&redirect_uri=http://localhost:3000/api/v1/callback/google&response_type=code&scope=https://www.googleapis.com/auth/drive.readonly https://www.googleapis.com/auth/userinfo.email&state=mock_state
	a.NotEmpty(decodedURL)
	a.Contains(decodedURL, "client_id="+gp.Config.ClientID)
	a.Contains(decodedURL, "redirect_uri="+gp.Config.RedirectURL)
	a.Contains(decodedURL, "response_type=code")
	a.Contains(decodedURL, "scope="+strings.Join(gp.Config.Scopes, " "))
	a.Contains(decodedURL, "state="+mockState)
}

func TestNewGoogleProvider_UpdateTokens(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	// Mock GoogleProvider
	gp := &GoogleProvider{}

	// Mock OAuth Token
	mockToken := &oauth2.Token{
		AccessToken:  "mock_access_token",
		RefreshToken: "mock_refresh_token",
		Expiry:       time.Now().Add(-100), // Expired Token
	}
	gp.Token = mockToken
	a.NotNil(gp.Token)

	err := gp.UpdateTokens(nil)
	a.Nil(err)
	a.Nil(gp.Token)

	// Mock GoogleAccount
	mockAccount := &types.GoogleAccount{
		ID:           "mock_id",
		UserID:       "mock_user_id",
		AccessToken:  "mock_access_token",
		RefreshToken: "mock_refresh_token",
		ExpiresAt:    time.Now().Add(time.Hour), // Mock expiry of one hour
		TokenType:    "Bearer",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err = gp.UpdateTokens(mockAccount)
	a.Nil(err)
	a.NotNil(gp.Token)
	a.Equal(gp.Token.AccessToken, mockAccount.AccessToken)
	a.Equal(gp.Token.RefreshToken, mockAccount.RefreshToken)
	a.Equal(gp.Token.Expiry, mockAccount.ExpiresAt)
	a.Equal(gp.Token.TokenType, mockAccount.TokenType)
}

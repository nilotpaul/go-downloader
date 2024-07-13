package store

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nilotpaul/go-downloader/config"
	"github.com/nilotpaul/go-downloader/service"
	"github.com/nilotpaul/go-downloader/types"
	"github.com/nilotpaul/go-downloader/util"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleProvider struct {
	Config     *oauth2.Config
	db         *sql.DB
	env        config.EnvConfig
	Token      *oauth2.Token
	HttpClient *http.Client
}

type googleProviderConfig struct {
	googleClientID     string
	googleClientSecret string
	googleRedirectURL  string
}

var scopes = []string{
	"https://www.googleapis.com/auth/drive.readonly",
	"https://www.googleapis.com/auth/userinfo.email",
}

func NewGoogleProvider(cfg googleProviderConfig, db *sql.DB, env config.EnvConfig) *GoogleProvider {
	config := &oauth2.Config{
		ClientID:     cfg.googleClientID,
		ClientSecret: cfg.googleClientSecret,
		RedirectURL:  cfg.googleRedirectURL,
		Scopes:       scopes,
		Endpoint:     google.Endpoint,
	}

	return &GoogleProvider{
		Config: config,
		db:     db,
		env:    env,
	}
}

func (g *GoogleProvider) Authenticate(authCode string) error {
	ctx := context.Background()
	token, err := g.Config.Exchange(ctx, authCode, oauth2.ApprovalForce)
	if err != nil {
		return util.NewAppError(
			http.StatusInternalServerError,
			"failed to authenticate",
			"NewGoogleProvider, Authenticate() error: ",
			err,
		)
	}

	if token == nil || !token.Valid() {
		return util.NewAppError(
			http.StatusBadRequest,
			"invalid oauth token",
			"NewGoogleProvider, Authenticate() error: ",
			err,
		)
	}

	g.Token = token
	g.HttpClient = g.Config.Client(ctx, token)

	return nil
}

func (g *GoogleProvider) GetAccessToken() string {
	if g.Token == nil {
		return ""
	}

	return g.Token.AccessToken
}

func (g *GoogleProvider) GetRefreshToken() string {
	if g.Token == nil {
		return ""
	}

	return g.Token.RefreshToken
}

func (g *GoogleProvider) IsTokenValid() bool {
	return g.Token.Valid()
}

func (g *GoogleProvider) RefreshToken(c *fiber.Ctx, userID string, force bool) (*oauth2.Token, error) {
	ctx := context.Background()
	if g.Token == nil {
		return nil, util.NewAppError(
			http.StatusNotFound,
			"no oauth token found",
			"NewGoogleProvider, RefreshToken() error",
		)
	}
	if force {
		g.Token.Expiry = time.Now().AddDate(-100, 0, 0)
	}
	tokenSrc := g.Config.TokenSource(ctx, g.Token)
	if tokenSrc == nil {
		return nil, util.NewAppError(
			http.StatusInternalServerError,
			"failed to generate refresh token",
			"NewGoogleProvider, RefreshToken() error",
		)
	}
	newToken, err := tokenSrc.Token()
	if err != nil {
		return nil, util.NewAppError(
			http.StatusInternalServerError,
			"failed to generate refresh token",
			"NewGoogleProvider, RefreshToken() error: ",
			err.Error(),
		)
	}

	g.Token = newToken
	g.HttpClient = g.Config.Client(ctx, newToken)
	if err := service.UpdateAccountByUserID(g.db, userID, &types.GoogleAccount{
		AccessToken:  g.Token.AccessToken,
		RefreshToken: g.Token.RefreshToken,
		TokenType:    g.Token.TokenType,
		ExpiresAt:    g.Token.Expiry,
	}); err != nil {
		return nil, util.NewAppError(
			http.StatusNotFound,
			"failed to update account with new tokens",
			"NewGoogleProvider, RefreshToken() error: ",
			err,
		)
	}

	return g.Token, nil
}

func (g *GoogleProvider) GetAuthURL(state string) string {
	return g.Config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
}

func (g *GoogleProvider) CreateOrUpdateAccount() (string, error) {
	// GetGoogleUserInfo uses the access token received during OAuth
	// and gets the user info from google.
	u, err := service.GetGoogleUserInfo(g.Token, g.HttpClient)
	if err != nil {
		return "", err
	}

	// GetUserByEmail gets the user row by email received
	// from GetGoogleUserInfo.
	dbUser, err := service.GetUserByEmail(g.db, u.Email)
	if err != nil {
		return "", util.NewAppError(
			http.StatusInternalServerError,
			"failed to retrieve user",
			"NewGoogleProvider, CreateOrUpdateAccount() error: ",
			err,
		)
	}
	// if user account already exists, we update the user's
	// account with new tokens and expiry.
	if len(dbUser.UserID) != 0 {
		acc := types.GoogleAccount{
			AccessToken:  g.Token.AccessToken,
			RefreshToken: g.Token.RefreshToken,
			TokenType:    g.Token.TokenType,
			ExpiresAt:    g.Token.Expiry,
		}
		if err := service.UpdateAccountByUserID(g.db, dbUser.UserID, &acc); err != nil {
			return "", util.NewAppError(
				http.StatusInternalServerError,
				"failed to update user",
				"NewGoogleProvider, CreateOrUpdateAccount() error: ",
				err,
			)
		}

		return dbUser.UserID, nil
	}
	// if user doesn't exists, we create new user account.
	userID, err := service.CreateUserAndAccount(g.db, u, g.Token)
	if err != nil {
		return "", util.NewAppError(
			http.StatusInternalServerError,
			"failed to create user and account",
			"NewGoogleProvider, CreateOrUpdateAccount() error: ",
			err,
		)
	}

	return userID, nil
}

func (g *GoogleProvider) CreateSession(c *fiber.Ctx, userID string) error {
	if g.Token == nil {
		return util.NewAppError(
			http.StatusNotFound,
			"no oauth token found",
			"NewGoogleProvider, RefreshToken() error",
		)
	}

	token, err := util.GenerateSessionToken(userID, g.env.SessionSecret)
	if err != nil {
		return util.NewAppError(
			http.StatusInternalServerError,
			"failed to create session",
			"NewGoogleProvider, RefreshToken() error: ",
			err,
		)
	}

	util.SetSessionToken(c, token, g.env.Domain)

	return nil
}

func (g *GoogleProvider) UpdateTokens(acc *types.GoogleAccount) error {
	if acc == nil {
		g.Token = nil
		return nil
	}

	g.Token = &oauth2.Token{
		AccessToken:  acc.AccessToken,
		RefreshToken: acc.RefreshToken,
		TokenType:    acc.TokenType,
		Expiry:       acc.ExpiresAt,
	}

	return nil
}

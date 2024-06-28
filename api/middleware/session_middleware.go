package middleware

import (
	"database/sql"
	"encoding/gob"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/nilotpaul/go-downloader/config"
	"github.com/nilotpaul/go-downloader/service"
	"github.com/nilotpaul/go-downloader/setting"
	"github.com/nilotpaul/go-downloader/store"
	"github.com/nilotpaul/go-downloader/types"
	"github.com/nilotpaul/go-downloader/util"
)

type SessionMiddleware struct {
	env       config.EnvConfig
	sessStore *session.Store
	db        *sql.DB
	registry  *store.ProviderRegistry
}

func NewSessionMiddleware(env config.EnvConfig, sessStore *session.Store, db *sql.DB, registry *store.ProviderRegistry) *SessionMiddleware {
	return &SessionMiddleware{
		env:       env,
		sessStore: sessStore,
		db:        db,
		registry:  registry,
	}
}

func init() {
	gob.Register(types.GoogleAccount{})
}

// SessionMiddleware will check if the validity of AccessToken token,
// if invalid it'll refresh the token and return the new credentions.
// values are create and updated in order to maintain the synchronicity.
func (m *SessionMiddleware) SessionMiddleware(c *fiber.Ctx) error {
	// GetSessionToken gets the jwt token from the cookie
	// which contains the UserID and JWT Expiry.
	token := util.GetSessionToken(c)
	if len(token) == 0 {
		slog.Error("SessionMiddleware", "error", "token length 0")
		return c.Next()
	}
	// verifies and extracts the UserID from JWT Token.
	decoded, err := util.VerifyAndDecodeSessionToken(c, token, m.env.SessionSecret)
	if err != nil {
		return util.NewAppError(
			http.StatusBadRequest,
			"invalid session",
			"SessionMiddleware error: ",
			err,
		)
	}

	session, err := service.GetAccountByUserID(m.db, decoded.UserID)
	if err != nil {
		return util.NewAppError(
			http.StatusInternalServerError,
			"invalid session",
			"SessionMiddleware error: ",
			err,
		)
	}
	if session == nil {
		return util.NewAppError(
			http.StatusNotFound,
			"no account found",
		)
	}

	gp, err := m.registry.GetProvider(setting.GoogleProvider)
	if err != nil {
		return util.NewAppError(
			http.StatusInternalServerError,
			"no provider found",
			"SessionMiddleware error: ",
			err,
		)
	}

	// injecting provider with the old and potentially (expired/invalid) tokens,
	// so that it can be used to refresh the token later.
	if err := gp.UpdateTokens(session); err != nil {
		return util.NewAppError(
			http.StatusInternalServerError,
			fmt.Errorf("failed to update the tokens: %s", err.Error()).Error(),
			"SessionMiddleware error: ",
			err,
		)
	}
	// if token has expired or is invalid, RefreshToken will generate a new
	// AccessToken and update the user account in database.
	if !gp.IsTokenValid() {
		t, err := gp.RefreshToken(c, session.UserId)
		if err != nil {
			return util.NewAppError(
				http.StatusInternalServerError,
				fmt.Errorf("failed to refresh the token: %s", err.Error()).Error(),
				"SessionMiddleware error: ",
				err,
			)
		}

		session.AccessToken = t.AccessToken
		session.RefreshToken = t.RefreshToken
		session.TokenType = t.TokenType
		session.ExpiresAt = t.Expiry
	}

	// injecting provider with the new tokens,
	if err := gp.UpdateTokens(session); err != nil {
		return util.NewAppError(
			http.StatusInternalServerError,
			"failed to update the tokens",
			"SessionMiddleware error: ",
			err,
		)
	}
	// setting the session in the in memory session store.
	if err := util.SetSessionInStore(c, m.sessStore, session); err != nil {
		return util.NewAppError(
			http.StatusInternalServerError,
			"failed to set the session in store",
			"SessionMiddleware error: ",
			err,
		)
	}

	return c.Next()
}

// WithGoogleOAuth will block access if the Token is invalid.
func (m *SessionMiddleware) WithGoogleOAuth(c *fiber.Ctx) error {
	gp, err := m.registry.GetProvider(setting.GoogleProvider)
	if err != nil {
		return util.NewAppError(
			http.StatusNotFound,
			"no provider found",
		)
	}

	if !gp.IsTokenValid() {
		return util.NewAppError(
			http.StatusUnauthorized,
			"invalid oauth token",
		)
	}

	return c.Next()
}

// WithGoogleOAuth will block access if the Token is valid.
func (m *SessionMiddleware) WithoutGoogleOAuth(c *fiber.Ctx) error {
	gp, err := m.registry.GetProvider(setting.GoogleProvider)
	if err != nil {
		return util.NewAppError(
			http.StatusNotFound,
			"no provider found",
		)
	}

	if gp.IsTokenValid() {
		return util.NewAppError(
			http.StatusUnauthorized,
			"your session is valid",
		)
	}

	return c.Next()
}

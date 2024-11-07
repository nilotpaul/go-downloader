package middleware

import (
	"database/sql"
	"encoding/gob"
	"log/slog"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
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
	gob.Register(types.GoogleAccountWrapper{})
}

// SessionMiddleware will check the validity of AccessToken token,
// if it's invalid it'll refresh the token and return the new credentials.
// values are created and updated in order to maintain the synchronicity.
func (m *SessionMiddleware) SessionMiddleware(c *fiber.Ctx) error {
	gp, err := m.registry.GetProvider(setting.GoogleProvider)
	if err != nil {
		return util.NewAppError(
			http.StatusInternalServerError,
			"no provider found",
			"SessionMiddleware error: ",
			err,
		)
	}

	// GetSessionToken gets the jwt token from the cookie
	// which contains the UserID and JWT Expiry.
	token := util.GetSessionToken(c)
	if len(token) == 0 {
		return c.Next()
	}

	// Verifies and extracts the UserID from JWT Token.
	decoded, err := util.VerifyAndDecodeSessionToken(token, m.env.SessionSecret)
	if err != nil {
		slog.Error("invalid session", "SessionMiddleware error", err)
		m.resetPersistingSession(c, gp)
		return c.Next()
	}

	session, err := service.GetAccountByUserID(m.db, decoded.UserID)
	if err != nil {
		slog.Error("invalid session", "SessionMiddleware error", err)
		m.resetPersistingSession(c, gp)
		return c.Next()
	}
	if session == nil {
		slog.Error("no account found", "SessionMiddleware error", err)
		m.resetPersistingSession(c, gp)
		return c.Next()
	}

	// Injecting provider with the old and potentially (expired/invalid) tokens,
	// so that it can be used to refresh the token later.
	if err := gp.UpdateTokens(session); err != nil {
		slog.Error("failed to update with old tokens", "SessionMiddleware error", err)
		m.resetPersistingSession(c, gp)
		return c.Next()
	}

	// If token has expired or is invalid, RefreshToken will generate a new
	// AccessToken and update the user account in database.
	if !gp.IsTokenValid() {
		t, err := gp.RefreshToken(c, session.UserID, false)
		if err != nil {
			slog.Error("failed to refresh the token", "SessionMiddleware error", err)
			m.resetPersistingSession(c, gp)
			return c.Next()
		}

		session.AccessToken = t.AccessToken
		session.RefreshToken = t.RefreshToken
		session.TokenType = t.TokenType
		session.ExpiresAt = t.Expiry
	}

	// Injecting provider with the new tokens.
	if err := gp.UpdateTokens(session); err != nil {
		slog.Error("failed to update with new tokens", "SessionMiddleware error", err)
		m.resetPersistingSession(c, gp)
		return c.Next()
	}

	// Setting the session in the in memory session store.
	if err := util.SetSessionInStore(c, m.sessStore, session); err != nil {
		slog.Error("failed to set the session in memory store", "SessionMiddleware error", err)
		m.resetPersistingSession(c, gp)
		return c.Next()
	}

	// Set the UserID in the request, already done in Session Storage
	// but this will be used in websocket conns as the type of
	// `fiber.Ctx `and `websocket.Conn` doesn't match, hence retrieving
	// sessions from storage is not possible.
	c.Locals(setting.LocalSessionKey, session.UserID)

	return c.Next()
}

// WithGoogleOAuth will block access if the Token is invalid.
func (m *SessionMiddleware) WithGoogleOAuth(c *fiber.Ctx) error {
	gp, err := m.registry.GetProvider(setting.GoogleProvider)
	if err != nil {
		return util.NewAppError(
			http.StatusUnauthorized,
			"no provider found",
		)
	}

	if !gp.IsTokenValid() {
		return util.NewAppError(
			http.StatusUnauthorized,
			"invalid session, please login",
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

// resetPersistingSession will update the session and token state with empty or nil values.
// It'll reset the persisting session when an error occurs in `SessionMiddleware`.
func (m *SessionMiddleware) resetPersistingSession(c *fiber.Ctx, gp types.OAuthProvider) {
	if err := gp.UpdateTokens(nil); err != nil {
		log.Error("failed reseting session(UpdateTokens): ", err)
	}
	if err := util.SetSessionInStore(c, m.sessStore, nil); err != nil {
		log.Error("failed reseting session(SetSessionInStore): ", err)
	}

	c.Locals(setting.LocalSessionKey, "")
}

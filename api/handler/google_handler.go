package handler

import (
	"database/sql"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/nilotpaul/go-downloader/config"
	"github.com/nilotpaul/go-downloader/service"
	"github.com/nilotpaul/go-downloader/setting"
	"github.com/nilotpaul/go-downloader/store"
	"github.com/nilotpaul/go-downloader/util"
)

type GoogleHandler struct {
	registry  *store.ProviderRegistry
	sessStore *session.Store
	db        *sql.DB
	env       config.EnvConfig
}

func NewGoogleHandler(registry *store.ProviderRegistry, sessStore *session.Store, db *sql.DB, env config.EnvConfig) *GoogleHandler {
	return &GoogleHandler{
		registry:  registry,
		sessStore: sessStore,
		db:        db,
		env:       env,
	}
}

// GoogleSignInHandler sends back an URL for the Google's consent page.
func (h *GoogleHandler) GoogleSignInHandler(c *fiber.Ctx) error {
	gp, err := h.registry.GetProvider(setting.GoogleProvider)
	if err != nil || gp == nil {
		return util.NewAppError(
			http.StatusNotFound,
			"provider not found",
		)
	}

	state, err := util.GenerateRandomState(32)
	if err != nil {
		return err
	}

	// GetAuthURL returns a URL to Google's consent page.
	authURL := gp.GetAuthURL(state)
	if len(authURL) == 0 {
		return util.NewAppError(
			http.StatusInternalServerError,
			"no authentication URL was generated",
		)
	}

	return c.JSON(fiber.Map{
		"url": authURL,
	})
}

// GoogleCallbackHandler uses `code` in URL and exchanges it with OAuth Token.
// Authenticates the user and create their session cookie.
func (h *GoogleHandler) GoogleCallbackHandler(c *fiber.Ctx) error {
	gp, err := h.registry.GetProvider(setting.GoogleProvider)
	if err != nil || gp == nil {
		return util.NewAppError(
			http.StatusNotFound,
			"provider not found",
		)
	}

	// Getting the `code` from the url.
	authCode := c.Query("code")
	if len(authCode) == 0 {
		return util.NewAppError(
			http.StatusBadRequest,
			"no authorization code found in URL",
		)
	}

	// Sets the access & refresh tokens in the GoogleProvider struct.
	if err := gp.Authenticate(authCode); err != nil {
		return err
	}

	// Creates or Updates the user account in database from the new tokens
	// that were inserted via Authenticate func above.
	userID, err := gp.CreateOrUpdateAccount()
	if err != nil {
		return err
	}

	// Creates and sets the session cookie.
	if err := gp.CreateSession(c, userID); err != nil {
		return err
	}

	// Redirecting the user to our App.
	return c.Redirect(util.GetEnv("REDIRECT_AFTER_LOGIN", "/"), http.StatusPermanentRedirect)
}

// GetSessionHandler sends back the user session.
func (h *GoogleHandler) GetSessionHandler(c *fiber.Ctx) error {
	// Gets the session from memory store.
	sess, err := util.GetSessionFromStore(c, h.sessStore)
	if err != nil {
		util.NewAppError(
			http.StatusUnauthorized,
			"no google oauth session found",
		)
	}

	// Fetching the user by `userID`.
	u, err := service.GetUserByID(h.db, sess.UserID)
	if err != nil {
		util.NewAppError(
			http.StatusInternalServerError,
			"failed to retieve the user session",
		)
	}
	if len(u.UserID) == 0 {
		util.NewAppError(
			http.StatusNotFound,
			"no google oauth session found",
		)
	}

	return c.JSON(u.Email)
}

// LogoutHandler clear the user session.
func (h *GoogleHandler) LogoutHandler(c *fiber.Ctx) error {
	gp, err := h.registry.GetProvider(setting.GoogleProvider)
	if err != nil {
		util.NewAppError(
			http.StatusNotFound,
			"no provider found",
		)
	}

	// Resets and clears the session state and cookie.
	err = util.ResetSession(c, gp, h.env.Domain)
	if err != nil {
		util.NewAppError(
			http.StatusNotFound,
			"failed to clear the session",
		)
	}

	return c.JSON("OK")
}

// RefreshTokenHandler forcefully refreshes the session even if it's still valid.
func (h *GoogleHandler) RefreshTokenHandler(c *fiber.Ctx) error {
	gp, err := h.registry.GetProvider(setting.GoogleProvider)
	if err != nil {
		return util.NewAppError(
			http.StatusNotFound,
			"no provider found",
			err,
		)
	}

	sess, err := util.GetSessionFromStore(c, h.sessStore)
	if err != nil {
		return util.NewAppError(
			http.StatusUnauthorized,
			"no session found",
			err,
		)
	}

	// Here, we forcefully refresh the session which might still be valid.
	t, err := gp.RefreshToken(c, sess.UserID, true)
	if err != nil {
		return err
	}

	sess.AccessToken = t.AccessToken
	sess.RefreshToken = t.RefreshToken
	sess.TokenType = t.TokenType
	sess.ExpiresAt = t.Expiry

	// Updating the token state.
	if err := gp.UpdateTokens(sess); err != nil {
		return util.NewAppError(
			http.StatusInternalServerError,
			"failed to update the tokens",
			err,
		)
	}
	// Updating the memory store.
	if err := util.SetSessionInStore(c, h.sessStore, sess); err != nil {
		return util.NewAppError(
			http.StatusInternalServerError,
			"failed to set the session in store",
			err,
		)
	}
	// Updating the local store (only for WS Connections).
	c.Locals(setting.LocalSessionKey, sess.UserID)

	return c.JSON("OK")
}

package handler

import (
	"database/sql"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/nilotpaul/go-downloader/service"
	"github.com/nilotpaul/go-downloader/setting"
	"github.com/nilotpaul/go-downloader/store"
	"github.com/nilotpaul/go-downloader/util"
)

type GoogleHandler struct {
	registry  *store.ProviderRegistry
	sessStore *session.Store
	db        *sql.DB
}

func NewGoogleHandler(registry *store.ProviderRegistry, sessStore *session.Store, db *sql.DB) *GoogleHandler {
	return &GoogleHandler{
		registry:  registry,
		sessStore: sessStore,
		db:        db,
	}
}

func (h *GoogleHandler) GoogleSignInHandler(c *fiber.Ctx) error {
	p, err := h.registry.GetProvider("google")
	if err != nil || p == nil {
		return util.NewAppError(
			http.StatusNotFound,
			"provider not found",
		)
	}

	state, err := util.GenerateRandomState(32)
	if err != nil {
		return err
	}

	authURL := p.GetAuthURL(state)
	if len(authURL) == 0 {
		return util.NewAppError(
			http.StatusInternalServerError,
			"no authentication URL was generated",
		)
	}

	return c.Redirect(authURL, http.StatusTemporaryRedirect)
}

func (h *GoogleHandler) GoogleCallbackHandler(c *fiber.Ctx) error {
	p, err := h.registry.GetProvider("google")
	if err != nil || p == nil {
		return fiber.NewError(http.StatusNotFound, err.Error())
	}

	authCode := c.Query("code")
	if len(authCode) == 0 {
		return util.NewAppError(
			http.StatusBadRequest,
			"no authorization code found in URL",
		)
	}

	// Sets the access & refresh tokens in the GoogleProvider struct.
	if err := p.Authenticate(authCode); err != nil {
		return err
	}

	userID, err := p.CreateOrUpdateAccount()
	if err != nil {
		return err
	}

	if err := p.CreateSession(c, userID); err != nil {
		return err
	}

	return c.Redirect(util.GetEnv("REDIRECT_AFTER_LOGIN", "/"), http.StatusPermanentRedirect)
}

func (h *GoogleHandler) GetSessionHandler(c *fiber.Ctx) error {
	sess, err := util.GetSessionFromStore(c, h.sessStore)
	if err != nil {
		util.NewAppError(
			http.StatusUnauthorized,
			"no google oauth session found",
		)
	}

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

func (h *GoogleHandler) LogoutHandler(c *fiber.Ctx) error {
	r, err := h.registry.GetProvider(setting.GoogleProvider)
	if err != nil {
		util.NewAppError(
			http.StatusNotFound,
			"no provider found",
		)
	}
	err = util.ResetSession(c, r)
	if err != nil {
		util.NewAppError(
			http.StatusNotFound,
			"failed to clear the session",
		)
	}

	return c.JSON("OK")
}

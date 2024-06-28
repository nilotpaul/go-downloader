package handler

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/nilotpaul/go-downloader/store"
	"github.com/nilotpaul/go-downloader/util"
)

type GoogleHandler struct {
	registry *store.ProviderRegistry
}

func NewGoogleHandler(registry *store.ProviderRegistry) *GoogleHandler {
	return &GoogleHandler{
		registry: registry,
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

func (h *GoogleHandler) LogoutHandler(c *fiber.Ctx) error {
	util.ResetSession(c)

	return c.JSON("OK")
}

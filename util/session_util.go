package util

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/nilotpaul/go-downloader/setting"
	"github.com/nilotpaul/go-downloader/types"
)

func GenerateSessionToken(userID string, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     setting.SessionExpiry.Unix(),
	})

	ts, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return ts, nil
}

func SetSessionToken(c *fiber.Ctx, token string, domain string) {
	c.Cookie(&fiber.Cookie{
		Name:     setting.SessionKey,
		Value:    token,
		Expires:  setting.SessionExpiry,
		HTTPOnly: true,
		Path:     "/",
		Secure:   false,
		Domain:   domain,
	})
}

func GetSessionToken(c *fiber.Ctx) string {
	return c.Cookies(setting.SessionKey, "")
}

func VerifyAndDecodeSessionToken(tokenStr string, secret string) (*types.JWTSession, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid signing method: %v", t.Method.Alg())
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid jwt claim")
	}
	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid userID")
	}
	expiry, ok := claims["exp"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid expires_at")
	}

	session := types.JWTSession{
		UserID:    userID,
		ExpiresAt: time.Unix(int64(expiry), 0),
	}

	return &session, nil
}

func GetSessionFromStore(c *fiber.Ctx, store *session.Store) (*types.GoogleAccount, error) {
	sess, err := store.Get(c)
	if err != nil {
		return nil, err
	}

	session := sess.Get(setting.SessionKey)
	if session == nil {
		return nil, fmt.Errorf("no session found")
	}
	v, ok := session.(types.GoogleAccountWrapper)
	if !ok {
		return nil, fmt.Errorf("invalid session type")
	}

	return v.GoogleAccount, nil
}

func SetSessionInStore(c *fiber.Ctx, store *session.Store, acc *types.GoogleAccount) error {
	sess, err := store.Get(c)
	if err != nil {
		return err
	}

	sess.Set(setting.SessionKey, types.GoogleAccountWrapper{
		GoogleAccount: acc,
	})
	if err := sess.Save(); err != nil {
		return err
	}

	return nil
}

func ResetSession(c *fiber.Ctx, r types.OAuthProvider, domain string) error {
	c.Cookie(&fiber.Cookie{
		Name:     setting.SessionKey,
		Path:     "/",
		HTTPOnly: true,
		Expires:  time.Now().AddDate(-100, 0, 0),
		Domain:   domain,
	})
	// clears the in memory session id cookie.
	c.Cookie(&fiber.Cookie{
		Name:     "session_id",
		HTTPOnly: true,
		Expires:  time.Now().AddDate(-100, 0, 0),
		Domain:   domain,
	})
	return r.UpdateTokens(nil)
}

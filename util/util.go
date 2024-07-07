package util

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"log/slog"
	"os"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/nilotpaul/go-downloader/setting"
)

type WebsocketFunc func(*websocket.Conn) error

func GetEnv(key string, fallback ...string) string {
	v := os.Getenv(key)
	if len(v) == 0 && len(fallback) > 0 {
		return fallback[0]
	}

	return v
}

func IsProduction() bool {
	e := GetEnv("ENVIRONMENT")

	return e == "PROD"
}

func DecodeJSON(r io.Reader, target interface{}) error {
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(target); err != nil {
		return err
	}

	return nil
}

// writeErrorResponse writes an error response to the WebSocket connection.
func writeErrorResponse(c *websocket.Conn, err error) {
	if appErr, ok := err.(*AppError); ok {
		slog.Error("WS error", "errMsg", appErr.Error(), "status", appErr.Status, "err", appErr.Err)
		err := c.WriteMessage(appErr.Status, []byte(appErr.Error()))
		if err != nil {
			log.Printf("failed to write error response: %v", err)
		}
	} else {
		slog.Error("WS error", "errMsg", appErr.Error(), "status", websocket.TextMessage, "err", "something went wrong")
		err := c.WriteMessage(websocket.TextMessage, []byte("something went wrong"))
		if err != nil {
			log.Printf("failed to write error response: %v", err)
		}
	}

	if err := c.Close(); err != nil {
		log.Printf("failed to close the ws connection: %v", err)
	}
}

// MakeWebsocketHandler creates a Fiber handler that wraps
// a WebSocket handler function.
func MakeWebsocketHandler(h WebsocketFunc) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return websocket.New(func(conn *websocket.Conn) {
				log.Println("new incoming ws connection", conn.NetConn().RemoteAddr())

				if err := h(conn); err != nil {
					writeErrorResponse(conn, err)
				}
			})(c)
		}
		return fiber.ErrUpgradeRequired
	}
}

func CommitOrRollback(tx *sql.Tx, err *error) {
	if *err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			*err = rbErr
		}
	} else {
		*err = tx.Commit()
	}
}

func MakeURL(url string) string {
	return setting.APIPrefix + url
}

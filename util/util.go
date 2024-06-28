package util

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"os"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
)

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

func CommitOrRollback(tx *sql.Tx, err *error) {
	if *err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			*err = rbErr
		}
	} else {
		*err = tx.Commit()
	}
}

func MakeTempl(c *fiber.Ctx, comp templ.Component) error {
	c.Set("Content-Type", "text/html")
	w := c.Response().BodyWriter()
	return comp.Render(context.Background(), w)
}

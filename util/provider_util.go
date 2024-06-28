package util

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

func GenerateRandomState(l int) (string, error) {
	b := make([]byte, l)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate a random state %s", err)
	}

	return base64.URLEncoding.EncodeToString(b), nil
}

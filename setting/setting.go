package setting

import (
	"time"
)

// supported providers
var Providers = []string{
	"google",
}

const (
	GoogleProvider = "google"
)

const (
	APIPrefix  string = "/api/v1"
	SessionKey string = "session_token"

	LocalSessionKey string = "session_user_id"
)

var SessionExpiry time.Time = time.Now().AddDate(0, 6, 0) // 6 months expiration time.

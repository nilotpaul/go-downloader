package setting

import (
	"time"
)

// supported providers
const (
	GoogleProvider = "google"
)

const SessionKey string = "session_token"

var SessionExpiry time.Time = time.Now().AddDate(0, 6, 0) // 6 months expiration time.

package setting

import (
	"time"
)

var (
	// Project folders to be excluded from folder tree overview.
	ExcludeDirs = []string{
		".vscode",
		".git",
		"api",
		"bin",
		"config",
		"dist",
		"migrations",
		"service",
		"setting",
		"store",
		"tmp",
		"types",
		"util",
	}
)

type Provider string

// Supported Providers List.
const (
	GoogleProvider Provider = "google"
)

// Other Utilities.
const (
	APIPrefix       string = "/api/v1"
	SessionKey      string = "session_token"
	LocalSessionKey string = "session_user_id"

	FolderPermission int = 0775
)

var SessionExpiry time.Time = time.Now().AddDate(0, 6, 0) // 6 months expiration time.

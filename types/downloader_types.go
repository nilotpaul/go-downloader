package types

import "time"

// `Progress` represents the state of a downloading file.
type Progress struct {
	FileID       string    `json:"file_id"`
	Total        int64     `json:"total"`
	Current      int       `json:"current"`
	Complete     bool      `json:"complete"`
	ReadableSize string    `json:"readableSize"`
	StartTime    time.Time `json:"startTime"`
	EndTime      time.Time `json:"endTime"`
	Speed        float64   `json:"speed"`
}

// `FolderNode` represents a node or folder in a hierarchical folder tree structure.
type FolderNode struct {
	Path     string       `json:"path"`
	Name     string       `json:"name"`
	Children []FolderNode `json:"children,omitempty"`
}

// Expected JSON Body data in download handler.
type DownloadHRBody struct {
	Links           string `json:"links"`
	DestinationPath string `json:"path"`
}

// Expected JSON Body data in cancel download handler.
type CancelDownloadHRBody struct {
	FileID string `json:"file_id"`
}

type Downloader interface {
	GetProgress(downloadingID string) (*Progress, error)
	SetProgress(downloadingID string, prog *Progress) error
}

package types

import "time"

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

type FolderNode struct {
	Path     string       `json:"path"`
	Name     string       `json:"name"`
	Children []FolderNode `json:"children,omitempty"`
}

type DownloadHRBody struct {
	Links           string `json:"links"`
	DestinationPath string `json:"path"`
}

type CancelDownloadHRBody struct {
	FileID string `json:"file_id"`
}

type Downloader interface {
	GetProgress(downloadingID string) (*Progress, error)
	SetProgress(downloadingID string, prog *Progress) error
}

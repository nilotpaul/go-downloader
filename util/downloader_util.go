package util

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v2"
	"google.golang.org/api/option"
)

const (
	_          = iota // ignore first value by assigning to blank identifier
	KB float64 = 1 << (10 * iota)
	MB
	GB
	TB
	PB
	EB
	ZB
	YB
)

func SanitizeFileName(fileName string) string {
	// Replace any invalid characters with an underscore
	regExp := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1F]`)
	newFileName := regExp.ReplaceAllString(fileName, "_")

	// Truncate the file name if it exceeds a certain length (255 chars)
	if len(newFileName) > 255 {
		newFileName = newFileName[:255]
	}

	return newFileName
}

func MakeGDriveService(accToken string) (*drive.Service, error) {
	token := &oauth2.Token{
		AccessToken: accToken,
	}
	ts := oauth2.StaticTokenSource(token)
	client := oauth2.NewClient(context.Background(), ts)

	srv, err := drive.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	return srv, nil
}

func CreateFile(path string) (*os.File, error) {
	dir, file := filepath.Split(path)

	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, err
	}

	f, err := os.Create(filepath.Join(dir, file))
	if err != nil {
		return nil, err
	}

	return f, nil
}

func GetGDriveFileID(url string) string {
	var fileID string
	exp1 := regexp.MustCompile(`drive\.google\.com\/open\?id\=(.*)`)
	exp2 := regexp.MustCompile(`drive\.google\.com\/file\/d\/(.*?)\/`)
	exp3 := regexp.MustCompile(`drive\.google\.com\/uc\?id\=(.*?)\&`)

	if matches := exp1.FindStringSubmatch(url); len(matches) > 1 {
		fileID = matches[1]
	} else if matches := exp2.FindStringSubmatch(url); len(matches) > 1 {
		fileID = matches[1]
	} else if matches := exp3.FindStringSubmatch(url); len(matches) > 1 {
		fileID = matches[1]
	} else {
		fileID = ""
	}

	return fileID
}

// FormatBytes returns a human-readable string representation
// of bytes in appropriate units
func FormatBytes(bytes int64) string {
	unit := ""
	size := float64(bytes)

	switch {
	case size >= YB:
		unit = "YB"
		size /= YB
	case size >= ZB:
		unit = "ZB"
		size /= ZB
	case size >= EB:
		unit = "EB"
		size /= EB
	case size >= PB:
		unit = "PB"
		size /= PB
	case size >= TB:
		unit = "TB"
		size /= TB
	case size >= GB:
		unit = "GB"
		size /= GB
	case size >= MB:
		unit = "MB"
		size /= MB
	case size >= KB:
		unit = "KB"
		size /= KB
	}

	return fmt.Sprintf("%.2f %s", size, unit)
}

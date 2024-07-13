package util

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/nilotpaul/go-downloader/types"
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

func sanitizeFolderPath(folderPath string) (string, bool) {
	// Replace any invalid characters with an underscore
	invalidCharsRegex := regexp.MustCompile(`[<>:"\\|?*\x00-\x1F]`)
	sanitizedPath := invalidCharsRegex.ReplaceAllString(folderPath, "_")

	wasValid := folderPath == sanitizedPath
	return sanitizedPath, wasValid
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

	if err := os.MkdirAll(dir, os.ModeDir|os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create the directories: %v", err)
	}

	f, err := os.Create(filepath.Join(dir, file))
	if err != nil {
		return nil, fmt.Errorf("failed to create the file: %v", err)
	}
	if err := os.Chmod(dir, os.ModeDir|os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to set directory permissions: %v", err)
	}
	if err := f.Chmod(os.FileMode(0664)); err != nil {
		return nil, fmt.Errorf("failed to set file permissions: %v", err)
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

func GetFileIDs(linksStr string) []string {
	var fileIds []string

	links := strings.Split(linksStr, ",")
	for _, link := range links {
		fileID := GetGDriveFileID(link)
		if len(fileID) != 0 {
			fileIds = append(fileIds, fileID)
		}
	}

	return fileIds
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

func ValidateDownloadHRBody(c *fiber.Ctx) (*types.DownloadHRBody, error) {
	var body types.DownloadHRBody
	if err := c.BodyParser(&body); err != nil {
		return nil, NewAppError(
			http.StatusUnprocessableEntity,
			"failed to parse the response body",
			err,
		)
	}

	_, valid := sanitizeFolderPath(body.DestinationPath)
	if len(body.DestinationPath) != 0 && !valid {
		return nil, NewAppError(
			http.StatusBadRequest,
			"invalid folder path",
		)
	}
	if len(body.Links) == 0 {
		return nil, NewAppError(
			http.StatusBadRequest,
			"invalid link(s)",
		)
	}

	return &body, nil
}

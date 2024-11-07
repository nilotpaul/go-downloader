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
	"github.com/nilotpaul/go-downloader/setting"
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

func MakeGDriveService(ctx context.Context, accToken string) (*drive.Service, error) {
	token := &oauth2.Token{
		AccessToken: accToken,
	}
	ts := oauth2.StaticTokenSource(token)
	client := oauth2.NewClient(ctx, ts)

	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
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

// GetGDriveFileID will extract the fileIDs from the url or link.
// The bool will return true for a file and vise-versa for a folder.
func GetGDriveFileID(url string) (string, bool) {
	var fileID string
	var isFile bool

	// for files
	expFile1 := regexp.MustCompile(`drive\.google\.com\/open\?id\=(.*)`)
	expFile2 := regexp.MustCompile(`drive\.google\.com\/file\/d\/([^\/\?]+)(\/|$)`)
	expFile3 := regexp.MustCompile(`drive\.google\.com\/uc\?id\=(.*?)(\/|$|&)`)

	// for folders
	expFolder1 := regexp.MustCompile(`drive\.google\.com\/drive\/folders\/([^\/\?]+)(\/|$|\?)`)
	expFolder2 := regexp.MustCompile(`drive\.google\.com\/drive\/u\/\d+\/folders\/([^\/\?]+)(\/|$|\?)`)
	expFolder3 := regexp.MustCompile(`drive\.google\.com\/folderview\?id\=(.*?)(\/|$|&)`)

	if matches := expFile1.FindStringSubmatch(url); len(matches) > 1 {
		fileID = matches[1]
		isFile = true
	} else if matches := expFile2.FindStringSubmatch(url); len(matches) > 1 {
		fileID = matches[1]
		isFile = true
	} else if matches := expFile3.FindStringSubmatch(url); len(matches) > 1 {
		fileID = matches[1]
		isFile = true
		// Check for folder IDs
	} else if matches := expFolder1.FindStringSubmatch(url); len(matches) > 1 {
		fileID = matches[1]
		isFile = false
	} else if matches := expFolder2.FindStringSubmatch(url); len(matches) > 1 {
		fileID = matches[1]
		isFile = false

	} else if matches := expFolder3.FindStringSubmatch(url); len(matches) > 1 {
		fileID = matches[1]
		isFile = false
	} else {
		fileID = ""
	}

	return fileID, isFile
}

func ParseGDriveIDs(linksStr string) map[string][]string {
	IDs := make(map[string][]string)

	links := strings.Split(linksStr, ",")
	for _, link := range links {
		fileID, isFile := GetGDriveFileID(link)
		if len(fileID) != 0 {
			if isFile {
				IDs["file"] = append(IDs["file"], fileID)
			} else {
				IDs["folder"] = append(IDs["folder"], fileID)
			}
		}
	}

	return IDs
}

func GetFileIDsFromFolder(srv *drive.Service, folderID string) ([]string, error) {
	var folderIDs []string
	query := fmt.Sprintf("'%s' in parents and trashed = false", folderID)
	pageToken := ""
	for {
		r, err := srv.Files.List().Q(query).PageToken(pageToken).MaxResults(100).Fields("nextPageToken, items(id, title)").Do()
		if err != nil {
			return folderIDs, err
		}
		for _, file := range r.Items {
			folderIDs = append(folderIDs, file.Id)
		}
		pageToken = r.NextPageToken
		if len(pageToken) == 0 {
			break
		}
	}

	return folderIDs, nil
}

func GetFolderTree(rootPath string) (*types.FolderNode, error) {
	var buildTree func(string) (*types.FolderNode, error)
	buildTree = func(path string) (*types.FolderNode, error) {
		info, err := os.Stat(path)
		if err != nil {
			return nil, err
		}

		// Skip files
		if !info.IsDir() {
			return nil, nil
		}

		// Skip excluded directories
		for _, excluded := range setting.ExcludeDirs {
			if excluded == path {
				return nil, nil
			}
		}

		node := &types.FolderNode{
			Path:     "./" + path,
			Name:     filepath.Base(path),
			Children: make([]types.FolderNode, 0),
		}
		files, err := os.ReadDir(path)
		if err != nil {
			return nil, err
		}

		for _, file := range files {
			childPath := filepath.Join(path, file.Name())
			childNode, err := buildTree(childPath)
			if err != nil {
				return nil, err
			}
			// Only add childNode to children if it's a directory
			if childNode != nil {
				node.Children = append(node.Children, *childNode)
			}
		}

		return node, nil
	}

	// Start building the tree from the rootPath
	rootNode, err := buildTree(rootPath)
	if err != nil {
		return nil, err
	}

	return rootNode, nil
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

func ValidateCancelDownloadHRBody(c *fiber.Ctx) (string, error) {
	var body types.CancelDownloadHRBody
	if err := c.BodyParser(&body); err != nil {
		return "", NewAppError(
			http.StatusUnprocessableEntity,
			"failed to parse the response body",
			err,
		)
	}

	if len(body.FileID) == 0 {
		return "", NewAppError(
			http.StatusBadRequest,
			"invalid fileID",
		)
	}

	return body.FileID, nil
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

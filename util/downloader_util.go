package util

import (
	"context"
	"os"
	"path/filepath"
	"regexp"

	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v2"
	"google.golang.org/api/option"
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

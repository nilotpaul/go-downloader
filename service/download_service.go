package service

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"math"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"github.com/nilotpaul/go-downloader/types"
	"github.com/nilotpaul/go-downloader/util"
)

type DownloaderConfig struct {
	FileID          string
	DestinationPath string
	FileName        string
	AccessToken     string
}

// GDriveDownloader will fallback to the original filename in if the `filename` parameter is an empty string.
func GDriveDownloader(cfg DownloaderConfig, progChan chan<- *types.Progress, ctx context.Context) error {
	// Validates the downloader configuration.
	if err := validateDownloaderConfig(cfg); err != nil {
		return err
	}

	// Making a GDrive Service with the access token from OAuth.
	srv, err := util.MakeGDriveService(ctx, cfg.AccessToken)
	if err != nil {
		return fmt.Errorf("failed to initialize GDrive service")
	}

	file, err := srv.Files.Get(cfg.FileID).Do()
	if err != nil {
		return fmt.Errorf("failed to get file info: %v", err)
	}

	// In case `fileID` is for a folder we return an error.
	if file.MimeType == "application/vnd.google-apps.folder" {
		return fmt.Errorf("expected file, received a folder")
	}

	// If no manual filename is provided, use the original filename from Google Drive
	if len(cfg.FileName) == 0 {
		cfg.FileName = file.OriginalFilename
	}

	// We take the destination path which is a folder location while the file will be downloaded.
	// Sanitize the filename to remove any invalid characters for file paths.
	destFileName := cfg.DestinationPath + "/" + util.SanitizeFileName(cfg.FileName)
	// Create the destination file, including any necessary directories.
	destFile, err := util.CreateFile(destFileName)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s", destFileName)
	}
	defer destFile.Close()

	res, err := srv.Files.Get(cfg.FileID).Download()
	if err != nil {
		return fmt.Errorf("failed with %d to download the file: %v", res.StatusCode, err)
	}
	defer res.Body.Close()

	// Creating a 32KB buffer which will hold a portion of the entire file for streaming.
	// Downloading the file in 32KB chunks is fast and memory efficient.
	buf := make([]byte, 32*1024) // 32KB buffer
	var totalWritten int64

	slog.Info("downloading", "filename", file.OriginalFilename)

	prog := &types.Progress{
		FileID:       cfg.FileID,
		Total:        file.FileSize,
		ReadableSize: util.FormatBytes(file.FileSize),
		StartTime:    time.Now(),
	}

	// Sending the initial progress
	progChan <- prog

	for {
		// Read the response body in chunks(32KB) and write it to the destination file,
		n, err := res.Body.Read(buf)

		// If `cancel` function is called from the download context, the loop will break
		// stopping the ongoing download.
		select {
		case <-ctx.Done():
			log.Infof("download cancelled for %s", cfg.FileID)
			return nil
		default:
			if n > 0 {
				written, writeErr := destFile.Write(buf[0:n])
				if writeErr != nil {
					return fmt.Errorf("failed to write the file content")
				}

				totalWritten += int64(written)
				prog.Current = int(float64(totalWritten) / float64(file.FileSize) * 100)
				elapsedTime := time.Since(prog.StartTime).Seconds()
				if elapsedTime > 0 {
					speed := ((float64(totalWritten) / elapsedTime) / 1e6) // Speed in Mbps
					prog.Speed = math.Round(speed*100) / 100               // Rounded to two decimal places
				}

				// Updating the downloading progress
				progChan <- prog
			}
		}

		if err != nil {
			// Break the loop if error is `EOF` -> End of Line which means the entire file has been downloaded.
			if err == io.EOF {
				break
			}
			// Otherwise break the loop and return with an error.
			return fmt.Errorf("failed to read response body of the file %s", file.OriginalFilename)
		}
	}

	// Mark download as complete
	prog.Complete = true
	prog.EndTime = time.Now()
	progChan <- prog

	return nil
}

func validateDownloaderConfig(cfg DownloaderConfig) error {
	if len(cfg.FileID) == 0 {
		return fmt.Errorf("invalid file id")
	}
	if len(cfg.DestinationPath) == 0 {
		return fmt.Errorf("invalid destination path")
	}
	if len(cfg.AccessToken) == 0 {
		return fmt.Errorf("invalid access token")
	}
	return nil
}

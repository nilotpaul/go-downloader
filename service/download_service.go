package service

import (
	"fmt"
	"io"
	"log/slog"
	"math"
	"time"

	"github.com/nilotpaul/go-downloader/types"
	"github.com/nilotpaul/go-downloader/util"
)

type DownloaderConfig struct {
	FileID          string
	UserID          string
	DestinationPath string
	FileName        string
	AccessToken     string
}

// GDriveDownloader will fallback to the original filename in if the filename parameter
// is an empty string.
func GDriveDownloader(cfg DownloaderConfig, progressChan chan<- *types.Progress) error {
	if err := validateDownloaderConfig(cfg); err != nil {
		return err
	}

	srv, err := util.MakeGDriveService(cfg.AccessToken)
	if err != nil {
		return fmt.Errorf("failed to initialize GDrive service")
	}

	file, err := srv.Files.Get(cfg.FileID).Do()
	if err != nil {
		return fmt.Errorf("failed to get file info: %v", err.Error())
	}
	if len(cfg.FileName) == 0 {
		cfg.FileName = file.OriginalFilename
	}

	destFileName := cfg.DestinationPath + "/" + util.SanitizeFileName(cfg.FileName)
	destFile, err := util.CreateFile(destFileName)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s", destFileName)
	}
	defer destFile.Close()

	res, err := srv.Files.Get(cfg.FileID).Download()
	if err != nil {
		return fmt.Errorf("failed with %d to download the file: %s", res.StatusCode, err.Error())
	}
	defer res.Body.Close()

	buf := make([]byte, 32*1024) // 32KB buffer
	var totalWritten int64

	slog.Info("downloading", "filename", file.OriginalFilename)

	prog := &types.Progress{
		FileID:       cfg.FileID,
		UserID:       cfg.UserID,
		Total:        file.FileSize,
		ReadableSize: util.FormatBytes(file.FileSize),
		StartTime:    time.Now(),
	}

	// Sending the initial progress
	progressChan <- prog

	for {
		n, err := res.Body.Read(buf)
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
			progressChan <- prog
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to read response body of the file %s", file.OriginalFilename)
		}
	}

	// Mark download as complete
	prog.Complete = true
	prog.EndTime = time.Now()
	progressChan <- prog

	return nil
}

func validateDownloaderConfig(cfg DownloaderConfig) error {
	if len(cfg.UserID) == 0 {
		return fmt.Errorf("invalid user id")
	}
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

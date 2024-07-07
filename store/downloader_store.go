package store

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/nilotpaul/go-downloader/service"
	"github.com/nilotpaul/go-downloader/types"
	"github.com/nilotpaul/go-downloader/util"
)

type Downloader struct {
	Progress        *sync.Map
	ProgresChan     chan *types.Progress
	FileIds         []string
	DestinationPath string
	UserID          string
}

func NewDownloader(fileIds []string, destinationPath string) *Downloader {
	return &Downloader{
		Progress:        &sync.Map{},
		ProgresChan:     make(chan *types.Progress),
		FileIds:         fileIds,
		DestinationPath: destinationPath,
	}
}

func (d *Downloader) StartDownload(accToken string, fileName string) error {
	for _, fileID := range d.FileIds {
		// Start multiple downloads in dedicated go routines
		go func(fileID string) {
			err := service.GDriveDownloader(service.DownloaderConfig{
				FileID:          fileID,
				UserID:          d.UserID,
				DestinationPath: d.DestinationPath,
				FileName:        fileName,
				AccessToken:     accToken,
			}, d.ProgresChan)
			if err != nil {
				fmt.Printf("Error downloading file %s: %v\n", fileID, err)
				return
			}
		}(fileID)
	}

	// Update the progress for each download
	go func() {
		for prog := range d.ProgresChan {
			d.Progress.Store(prog.FileID, prog)
		}
	}()

	return nil
}

func (d *Downloader) GetProgress(fileID string, userID string) (*types.Progress, error) {
	v, ok := d.Progress.Load(fileID)
	if !ok {
		return nil, util.NewAppError(
			http.StatusNotFound,
			fmt.Sprintf("no current downloads for this ID %s", fileID),
			"Downloader, GetProgress error",
		)
	}

	prog, ok := v.(*types.Progress)
	if !ok {
		return nil, util.NewAppError(
			http.StatusInternalServerError,
			fmt.Sprintf("invalid progress type for file ID %s", fileID),
			"Downloader, GetProgress error",
		)
	}
	if prog.UserID != userID {
		return nil, util.NewAppError(
			http.StatusNotFound,
			fmt.Sprintf("no current downloads for this ID %s", fileID),
			"Downloader, GetProgress error: progress and locals userID didn't matched",
		)
	}

	return prog, nil
}

func (d *Downloader) SetProgress(fileID string, prog *types.Progress) {
	d.Progress.Store(fileID, prog)
}

func (d *Downloader) UpdateProgress(progress *types.Progress) {
	d.Progress.Store(progress.FileID, progress)
}

func (d *Downloader) DeleteProgress(fileID string) {
	d.Progress.Delete(fileID)
}

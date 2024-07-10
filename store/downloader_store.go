package store

import (
	"fmt"
	"sync"

	"github.com/gofiber/fiber/v2/log"
	"github.com/nilotpaul/go-downloader/service"
	"github.com/nilotpaul/go-downloader/types"
)

type Downloader struct {
	progressChans      map[string]chan *types.Progress
	ErrChans           map[string]chan error
	FileIds            []string
	DestinationPath    string
	UserID             string
	PendingDownloads   map[string]*types.Progress
	pendingDownloadsMu sync.RWMutex
}

func NewDownloader(fileIds []string, destinationPath string) *Downloader {
	return &Downloader{
		progressChans:    make(map[string]chan *types.Progress),
		ErrChans:         make(map[string]chan error),
		FileIds:          fileIds,
		DestinationPath:  destinationPath,
		PendingDownloads: make(map[string]*types.Progress),
	}
}

func (d *Downloader) StartDownload(accToken string, fileName string) error {
	for _, fileID := range d.FileIds {

		progChan := make(chan *types.Progress)
		d.progressChans[fileID] = progChan

		errChan := make(chan error)
		d.ErrChans[fileID] = errChan

		// Start multiple downloads in dedicated go routines
		go func(fileID string) {
			err := service.GDriveDownloader(service.DownloaderConfig{
				FileID:          fileID,
				UserID:          d.UserID,
				DestinationPath: d.DestinationPath,
				FileName:        fileName,
				AccessToken:     accToken,
			}, progChan)
			if err != nil {
				log.Errorf("error downloading file %s: %v\n ", fileID, err)
				errChan <- err

				d.cleanUp(fileID, progChan)
				return
			}
		}(fileID)
	}

	for fileID, progChan := range d.progressChans {
		go func(fileID string, progChan chan *types.Progress) {
			for prog := range progChan {
				d.pendingDownloadsMu.Lock()
				d.PendingDownloads[fileID] = prog
				d.pendingDownloadsMu.Unlock()

				if prog.Complete {
					d.cleanUp(fileID, progChan)
					break
				}
			}
		}(fileID, progChan)
	}

	return nil
}

func (d *Downloader) GetPendingDownloads(userID string) ([]*types.Progress, error) {
	d.pendingDownloadsMu.Lock()
	defer d.pendingDownloadsMu.Unlock()

	if d.PendingDownloads == nil {
		return nil, fmt.Errorf("no ongoing downloads")
	}

	var pendingsDownloads []*types.Progress
	for _, prog := range d.PendingDownloads {
		// Check if the progress belongs to the specified userID
		if prog.UserID == userID {
			pendingsDownloads = append(pendingsDownloads, prog)
		}
	}

	return pendingsDownloads, nil
}

func (d *Downloader) GetProgress(fileID string, userID string) (*types.Progress, error) {
	d.pendingDownloadsMu.Lock()
	defer d.pendingDownloadsMu.Unlock()

	if d.PendingDownloads == nil {
		return nil, fmt.Errorf("no ongoing downloads")
	}

	prog := d.PendingDownloads[fileID]
	if prog.UserID != userID {
		log.Info("Downloader, GetProgress: the user initiated the download and the one requesting the progress doesn't match")
		return nil, fmt.Errorf("no ongoing downloads")
	}

	return prog, nil
}

func (d *Downloader) SetProgress(fileID string, prog *types.Progress) {
	d.pendingDownloadsMu.Lock()
	defer d.pendingDownloadsMu.Unlock()

	if existingProg, ok := d.PendingDownloads[fileID]; ok {
		existingProg.Complete = prog.Complete
		existingProg.Current = prog.Current
		existingProg.EndTime = prog.EndTime
	} else {
		d.PendingDownloads[fileID] = prog
	}
}

func (d *Downloader) DeleteProgress(fileID string) {
	d.pendingDownloadsMu.Lock()
	defer d.pendingDownloadsMu.Unlock()

	delete(d.PendingDownloads, fileID)
}

func (d *Downloader) cleanUp(fileID string, progChan chan *types.Progress) {
	d.pendingDownloadsMu.Lock()
	delete(d.PendingDownloads, fileID)
	d.pendingDownloadsMu.Unlock()

	close(progChan)
	delete(d.progressChans, fileID)
	close(d.ErrChans[fileID])
	delete(d.ErrChans, fileID)
}

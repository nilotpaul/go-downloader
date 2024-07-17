package store

import (
	"context"
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
	PendingDownloads   map[string]*types.Progress
	cancelFuncs        map[string]context.CancelFunc
	pendingDownloadsMu sync.RWMutex
}

func NewDownloader(fileIds []string, destinationPath string) *Downloader {
	return &Downloader{
		progressChans:    make(map[string]chan *types.Progress),
		ErrChans:         make(map[string]chan error),
		FileIds:          fileIds,
		DestinationPath:  destinationPath,
		PendingDownloads: make(map[string]*types.Progress),
		cancelFuncs:      make(map[string]context.CancelFunc),
	}
}

func (d *Downloader) StartDownload(ctx context.Context, accToken string, fileName string) error {
	for _, fileID := range d.FileIds {
		progChan := make(chan *types.Progress)
		d.progressChans[fileID] = progChan

		errChan := make(chan error)
		d.ErrChans[fileID] = errChan

		downloadCtx, cancel := context.WithCancel(ctx)
		d.cancelFuncs[fileID] = cancel

		// Start multiple downloads in dedicated go routines
		go func(fileID string) {
			err := service.GDriveDownloader(service.DownloaderConfig{
				FileID:          fileID,
				DestinationPath: d.DestinationPath,
				FileName:        fileName,
				AccessToken:     accToken,
			}, progChan, downloadCtx)
			if err != nil {
				log.Errorf("error downloading file %s: %v\n", fileID, err)
				errChan <- err
			}

			d.cleanUp(fileID)
		}(fileID)
	}

	for fileID, progChan := range d.progressChans {
		go func(fileID string, progChan chan *types.Progress) {
			for prog := range progChan {
				d.SetProgress(fileID, prog)
			}
		}(fileID, progChan)
	}

	return nil
}

func (d *Downloader) GetPendingDownloads() ([]*types.Progress, error) {
	d.pendingDownloadsMu.Lock()
	defer d.pendingDownloadsMu.Unlock()

	if d.PendingDownloads == nil {
		return nil, fmt.Errorf("no ongoing downloads")
	}

	var pendingsDownloads []*types.Progress
	for _, prog := range d.PendingDownloads {
		// Check if the progress belongs to the specified userID
		pendingsDownloads = append(pendingsDownloads, prog)
	}

	return pendingsDownloads, nil
}

func (d *Downloader) GetProgress(fileID string) (*types.Progress, error) {
	d.pendingDownloadsMu.Lock()
	defer d.pendingDownloadsMu.Unlock()

	if d.PendingDownloads == nil {
		return nil, fmt.Errorf("no ongoing downloads")
	}

	prog := d.PendingDownloads[fileID]

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

func (d *Downloader) CancelDownload(fileID string) error {
	cancel, ok := d.cancelFuncs[fileID]
	if !ok {
		return fmt.Errorf("no downloads found to cancel")
	}
	cancel()

	return nil
}

func (d *Downloader) CancelAllDownloads() {
	for _, cancel := range d.cancelFuncs {
		cancel()
	}
}

func (d *Downloader) cleanUp(fileID string) {
	close(d.progressChans[fileID])
	close(d.ErrChans[fileID])

	delete(d.progressChans, fileID)
	delete(d.ErrChans, fileID)

	d.pendingDownloadsMu.Lock()
	delete(d.PendingDownloads, fileID)
	d.pendingDownloadsMu.Unlock()
}

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
	FileIDs            []string
	DestinationPath    string
	PendingDownloads   map[string]*types.Progress
	cancelFuncs        map[string]context.CancelFunc
	pendingDownloadsMu sync.RWMutex
}

func NewDownloader(fileIds []string, destinationPath string) *Downloader {
	return &Downloader{
		progressChans:    make(map[string]chan *types.Progress),
		ErrChans:         make(map[string]chan error),
		FileIDs:          fileIds,
		DestinationPath:  destinationPath,
		PendingDownloads: make(map[string]*types.Progress),
		cancelFuncs:      make(map[string]context.CancelFunc),
	}
}

func (d *Downloader) StartDownload(ctx context.Context, accToken string, fileName string) error {
	// For every file
	for _, fileID := range d.FileIDs {
		// Making progress channel and storing it in the `progressChans` map.
		progChan := make(chan *types.Progress)
		d.progressChans[fileID] = progChan

		// Making error channel and storing it in the `ErrChans` map.
		errChan := make(chan error)
		d.ErrChans[fileID] = errChan

		// Making context for each file and storing it in `cancelFuncs` map.
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

			// For any errors in between download, they'll be sent to their respective error channel.
			if err != nil {
				log.Errorf("error downloading file %s: %v\n", fileID, err)
				errChan <- err
			}

			// Closing the progress and error channel, deleting them from `progressChans` and `errorChans` map
			// with it's progress status to mark the download as complete -> can be due
			// to an error or successful completion.
			d.cleanUp(fileID)
		}(fileID)
	}

	// We create dedicated go routines to handle progress updates for each file.
	for fileID, progChan := range d.progressChans {
		go d.handleProgressUpdates(fileID, progChan)
	}

	return nil
}

// handleProgressUpdates takes a `fileID` and `progChan`, ranges over the channel itself
// and sets it's progress continuously.
func (d *Downloader) handleProgressUpdates(fileID string, progChan chan *types.Progress) {
	for prog := range progChan {
		d.SetProgress(fileID, prog)
	}
}

// GetPendingDownloads a slice of `progress` which is a pointer, it ranges
// over the `pendingsDownloads` map and retrieves all the downloads from `PendingDownloads` map
func (d *Downloader) GetPendingDownloads() ([]*types.Progress, error) {
	d.pendingDownloadsMu.Lock()
	defer d.pendingDownloadsMu.Unlock()

	if d.PendingDownloads == nil {
		return nil, fmt.Errorf("no ongoing downloads")
	}

	var pendingsDownloads []*types.Progress
	for _, prog := range d.PendingDownloads {
		pendingsDownloads = append(pendingsDownloads, prog)
	}

	return pendingsDownloads, nil
}

// GetProgress retrieves the current progress for a file by it's fileID.
func (d *Downloader) GetProgress(fileID string) (*types.Progress, error) {
	d.pendingDownloadsMu.Lock()
	defer d.pendingDownloadsMu.Unlock()

	if d.PendingDownloads == nil {
		return nil, fmt.Errorf("no ongoing downloads")
	}

	prog := d.PendingDownloads[fileID]

	return prog, nil
}

// SetProgress sets the progress for a file by it's fileID.
func (d *Downloader) SetProgress(fileID string, prog *types.Progress) {
	d.pendingDownloadsMu.Lock()
	defer d.pendingDownloadsMu.Unlock()

	// If the progress already exists in the `PendingDownloads` map then update it.
	if existingProg, ok := d.PendingDownloads[fileID]; ok {
		existingProg.Complete = prog.Complete
		existingProg.Current = prog.Current
		existingProg.EndTime = prog.EndTime
	} else {
		// If the progress doesn't exist in the `PendingDownloads` map then create it.
		d.PendingDownloads[fileID] = prog
	}
}

// DeleteProgress removes the progress for a file from `PendingDownloads` map.
func (d *Downloader) DeleteProgress(fileID string) {
	d.pendingDownloadsMu.Lock()
	defer d.pendingDownloadsMu.Unlock()

	delete(d.PendingDownloads, fileID)
}

// CancelDownload gets the cancel function for the download context from
// `cancelFuncs` map by `fileID` and executes it which stops the ongoing download.
func (d *Downloader) CancelDownload(fileID string) error {
	cancel, ok := d.cancelFuncs[fileID]
	if !ok {
		return fmt.Errorf("no downloads found to cancel")
	}
	cancel()

	return nil
}

// CancelAllDownloads ranges over the `cancelFuncs` and executes each `cancel` functions
// to cancel all ongoing downloads.
func (d *Downloader) CancelAllDownloads() {
	for _, cancel := range d.cancelFuncs {
		cancel()
	}
}

// cleanUp removes the state for a file, progress and error channels are closed and deleted,
// progress status is also removed.
func (d *Downloader) cleanUp(fileID string) {
	close(d.progressChans[fileID])
	close(d.ErrChans[fileID])

	delete(d.progressChans, fileID)
	delete(d.ErrChans, fileID)

	d.pendingDownloadsMu.Lock()
	delete(d.PendingDownloads, fileID)
	d.pendingDownloadsMu.Unlock()
}

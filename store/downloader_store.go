package store

import (
	"context"
	"fmt"
	"sync"

	"github.com/gofiber/fiber/v2/log"
	"github.com/nilotpaul/go-downloader/service"
	"github.com/nilotpaul/go-downloader/types"
)

// Will be used later when we refactor for supporting multiple providers.
// type downloadFunc func(interface{}, context.Context, string) error

// Note: This is not a generic function which can support all provider.
type downloadFunc func(service.GDriveDownloadConfig, chan<- *types.Progress, context.Context) error

type Downloader struct {
	// progressChans map holds progress channel for every file to be downloaded.
	progressChans map[string]chan *types.Progress

	// ErrChans map holds the error channel for every file to be downloaded.
	ErrChans map[string]chan error

	// FileIDs is a list of all files that needs to be downloaded.
	FileIDs []string

	// DestinationPath is the location where we want to store downloads.
	// This can be an empty string representing that a default path will
	// be used provided in env.
	DestinationPath string

	// PendingDownloads map holds the progress status for every ongoing
	// downloads aggregated.
	PendingDownloads map[string]*types.Progress
	// pendingDownloadsMu is used for thread safe access of PendingDownloads map.
	pendingDownloadsMu sync.RWMutex

	// downloadFunc is a generic function used for downloading files.
	downloadFunc downloadFunc

	// cancelFuncs map holds cancel function from `context.WithCancel(ctx)` for
	// every file that needs to be downloaded.
	cancelFuncs map[string]context.CancelFunc
}

func NewDownloader(fileIds []string, destinationPath string) *Downloader {
	return &Downloader{
		progressChans:    make(map[string]chan *types.Progress),
		ErrChans:         make(map[string]chan error),
		FileIDs:          fileIds,
		DestinationPath:  destinationPath,
		PendingDownloads: make(map[string]*types.Progress),
		cancelFuncs:      make(map[string]context.CancelFunc),
		downloadFunc:     service.GDriveDownloader,
	}
}

func (d *Downloader) StartDownload(ctx context.Context, accToken string, fileName string) error {
	// For every file
	for _, fileID := range d.FileIDs {
		// Initializing the necessary states for each file and
		// storing them for later usage.
		state := d.initializeDownload(fileID, ctx)

		// Config for a single file
		downloadCfg := service.DownloaderConfig{
			FileID:          fileID,
			DestinationPath: d.DestinationPath,
			FileName:        fileName,
		}

		// Start multiple downloads in dedicated go routines
		go func(fileID string) {
			err := d.downloadFunc(service.GDriveDownloadConfig{
				DownloaderConfig: downloadCfg,
				AccessToken:      accToken,
			}, state.progChan, state.downloadCtx)

			// For any errors in between download, they'll be sent to their respective error channel.
			if err != nil {
				log.Errorf("error downloading file %s: %v\n", fileID, err)
				state.errChan <- err
			}

			// Closing the progress and error channel, deleting them from `progressChans` and `errorChans` map
			// with it's progress status to mark the download as complete -> can be due
			// to an error or successful completion.
			// time.Sleep(300 * time.Second)
			d.cleanUp(fileID)
		}(fileID)
	}

	// We create dedicated go routines to handle progress updates for each file.
	for fileID, progChan := range d.progressChans {
		go d.handleProgressUpdates(fileID, progChan)
	}

	return nil
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

// State that will be initialized by initializeDownload below.
type initializedState struct {
	progChan    chan *types.Progress
	errChan     chan error
	downloadCtx context.Context
}

// initializeDownload creates necessary states for a file to be downloaded and stores them
// in the `Download` struct and returns the `initializedState`.
func (d *Downloader) initializeDownload(fileID string, ctx context.Context) initializedState {
	// Making progress channel and storing it in the `progressChans` map.
	progChan := make(chan *types.Progress)
	d.progressChans[fileID] = progChan

	// Making error channel and storing it in the `ErrChans` map.
	errChan := make(chan error)
	d.ErrChans[fileID] = errChan

	// Making context for each file and storing it in `cancelFuncs` map.
	downloadCtx, cancel := context.WithCancel(ctx)
	d.cancelFuncs[fileID] = cancel

	return initializedState{
		progChan:    progChan,
		errChan:     errChan,
		downloadCtx: downloadCtx,
	}
}

// handleProgressUpdates takes a `fileID` and `progChan`, ranges over the channel itself
// and sets it's progress continuously.
func (d *Downloader) handleProgressUpdates(fileID string, progChan chan *types.Progress) {
	for prog := range progChan {
		d.SetProgress(fileID, prog)
	}
}

// cleanUp removes the state for a file, progress and error channels are closed and deleted,
// progress status is also removed.
func (d *Downloader) cleanUp(fileID string) {
	d.pendingDownloadsMu.Lock()
	delete(d.PendingDownloads, fileID)
	d.pendingDownloadsMu.Unlock()

	close(d.ErrChans[fileID])
	close(d.progressChans[fileID])

	delete(d.progressChans, fileID)
	delete(d.ErrChans, fileID)
}

package service

import (
	"io"
	"net/http"

	"github.com/nilotpaul/go-downloader/util"
)

// GDriveDownloader will fallback to the original filename in GDrive if the filename
// parameter is an empty string.
func GDriveDownloader(fileID string, destPath string, filename string, accToken string) error {
	srv, err := util.MakeGDriveService(accToken)
	if err != nil {
		return util.NewAppError(
			http.StatusInternalServerError,
			"failed to initialize GDrive service",
			"GDriveDownloader error: ",
			err,
		)
	}

	file, err := srv.Files.Get(fileID).Do()
	if err != nil {
		return util.NewAppError(
			http.StatusInternalServerError,
			err.Error(),
			"GDriveDownloader error: ",
			err,
		)
	}
	if len(filename) == 0 {
		filename = file.OriginalFilename
	}

	destFile, err := util.CreateFile(destPath + "/" + util.SanitizeFileName(filename))
	if err != nil {
		return util.NewAppError(
			http.StatusInternalServerError,
			"failed to create destination path",
			"GDriveDownloader error: ",
			err,
		)
	}
	defer destFile.Close()

	res, err := srv.Files.Get(fileID).Download()
	if err != nil {
		return util.NewAppError(
			http.StatusInternalServerError,
			"failed to download the file",
			"GDriveDownloader error: ",
			err,
		)
	}
	defer res.Body.Close()

	_, err = io.Copy(destFile, res.Body)
	if err != nil {
		return util.NewAppError(
			http.StatusInternalServerError,
			"failed to write the file content",
			"GDriveDownloader error: ",
			err,
		)
	}

	return nil
}

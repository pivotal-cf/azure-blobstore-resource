package api

import (
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/mholt/archiver"
)

type In struct {
	azureClient azureClient
}

func NewIn(azureClient azureClient) In {
	return In{
		azureClient: azureClient,
	}
}

func (i In) CopyBlobToDestination(destinationDir, blobName string, snapshot *time.Time, blockSize int64, retryTryTimeout time.Duration) error {
	fileName := path.Base(blobName)
	file, err := os.Create(filepath.Join(destinationDir, fileName))
	if err != nil {
		return err
	}
	defer file.Close()

	return i.azureClient.DownloadBlobToFile(blobName, file, snapshot, blockSize, retryTryTimeout)
}

func (i In) UnpackBlob(filename string) error {
	err := archiver.Unarchive(filename, filepath.Dir(filename))
	if err != nil {
		return err
	}

	err = os.Remove(filename)
	if err != nil {
		// not tested
		return err
	}

	return nil
}

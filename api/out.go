package api

import (
	"os"
	"path/filepath"
	"time"
)

type Out struct {
	azureClient azureClient
}

func NewOut(azureClient azureClient) Out {
	return Out{
		azureClient: azureClient,
	}
}

func (o Out) UploadFileToBlobstore(sourceDirectory string, filename string, blobName string) (time.Time, error) {
	file, err := os.Open(filepath.Join(sourceDirectory, filename))
	if err != nil {
		return time.Time{}, err
	}
	defer file.Close()

	err = o.azureClient.UploadFromStream(blobName, file)
	if err != nil {
		return time.Time{}, err
	}

	snapshot, err := o.azureClient.CreateSnapshot(blobName)
	if err != nil {
		return time.Time{}, err
	}

	return snapshot, nil
}

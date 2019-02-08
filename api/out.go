package api

import (
	"fmt"
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

func (o Out) UploadFileToBlobstore(sourceDirectory string, filename string, blobName string, createSnapshot bool) (string, time.Time, error) {
	matches, err := filepath.Glob(filepath.Join(sourceDirectory, filename))
	if err != nil {
		// not tested
		return "", time.Time{}, err
	}

	var fileToUpload string
	if len(matches) == 0 {
		fileToUpload = filepath.Join(sourceDirectory, filename)
	} else if len(matches) == 1 {
		fileToUpload = matches[0]
		if !createSnapshot {
			blobName = filepath.Join(filepath.Dir(blobName), filepath.Base(fileToUpload))
		}
	} else {
		return "", time.Time{}, fmt.Errorf("multiple files match glob: %s", filename)
	}

	file, err := os.Open(fileToUpload)
	if err != nil {
		return "", time.Time{}, err
	}
	defer file.Close()

	err = o.azureClient.UploadFromStream(blobName, file)
	if err != nil {
		return "", time.Time{}, err
	}

	if createSnapshot {
		snapshot, err := o.azureClient.CreateSnapshot(blobName)
		if err != nil {
			return "", time.Time{}, err
		}

		return blobName, snapshot, nil
	}

	return blobName, time.Time{}, nil
}

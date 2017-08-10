package api

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

type In struct {
	azureClient azureClient
}

func NewIn(azureClient azureClient) In {
	return In{
		azureClient: azureClient,
	}
}

func (i In) CopyBlobToDestination(destinationDir, blobName string, snapshot time.Time) error {
	data, err := i.azureClient.Get(blobName, snapshot)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(destinationDir, blobName), data, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

package api

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"errors"
	"fmt"
)

type In struct {
	azureClient azureClient
}

func NewIn(azureClient azureClient) In {
	return In{
		azureClient: azureClient,
	}
}

// Downloads a blob to a target filepath location.
func (i In) CopyBlobToDestination(destinationDir, blobName string) error {
	data, err := i.azureClient.Get(blobName)
	if err != nil {
		return errors.New(fmt.Sprintf("failed to get target blob: %s", err))
	}

	targetPath := filepath.Join(destinationDir, blobName)
	err = ioutil.WriteFile(targetPath, data, os.ModePerm)
	if err != nil {
		return errors.New(fmt.Sprintf("failed to write target blob to %s: %s", targetPath, err))
	}

	return nil
}

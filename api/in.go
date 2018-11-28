package api

import (
	"io"
	"os"
	"path"
	"path/filepath"
	"time"
)

const (
	ChunkSize = 4000000 // 4Mb
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
	blobSize, err := i.azureClient.GetBlobSizeInBytes(blobName, snapshot)
	if err != nil {
		return err
	}

	subDir := path.Dir(blobName)
	if subDir != "" {
		err := os.MkdirAll(filepath.Join(destinationDir, subDir), os.ModePerm)
		if err != nil {
			return err
		}
	}

	file, err := os.Create(filepath.Join(destinationDir, blobName))
	if err != nil {
		return err
	}
	defer file.Close()

	var start, end uint64
	end = min(ChunkSize-1, uint64(blobSize))

	for start < uint64(blobSize) {
		blobReader, err := i.azureClient.GetRange(blobName, start, end, snapshot)
		if err != nil {
			return err
		}

		_, err = io.Copy(file, blobReader)
		if err != nil {
			return err
		}

		blobReader.Close()

		start = min(start+ChunkSize, uint64(blobSize))
		end = min(end+ChunkSize, uint64(blobSize))
	}

	return nil
}

func min(x, y uint64) uint64 {
	if x < y {
		return x
	}
	return y
}

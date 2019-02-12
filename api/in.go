package api

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
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

func (i In) UnpackBlob(filename string) error {
	fileExtension := filepath.Ext(filename)
	var cmd *exec.Cmd

	switch fileExtension {
	case ".gz":
		cmd = exec.Command("gzip", "-d", filename)
	case ".tgz":
		cmd = exec.Command("tar", "-xvf", filename, "-C", filepath.Dir(filename))
	case ".zip":
		cmd = exec.Command("unzip", filename, "-d", filepath.Dir(filename))
	default:
		return fmt.Errorf("invalid extension: %s", filename)
	}

	var out bytes.Buffer
	cmd.Stderr = &out
	err := cmd.Run()

	if err != nil {
		return errors.New(out.String())
	}

	return nil
}

func min(x, y uint64) uint64 {
	if x < y {
		return x
	}
	return y
}

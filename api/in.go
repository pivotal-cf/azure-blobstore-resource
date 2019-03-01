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
	"strings"
	"time"

	"github.com/h2non/filetype"
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
	var cmd *exec.Cmd

	fileType, err := mimeType(filename)
	if err != nil {
		return err
	}

	switch fileType {
	case "application/gzip":
		cmd = exec.Command("gzip", "-d", filename)
	case "application/x-tar":
		cmd = exec.Command("tar", "-xvf", filename, "-C", filepath.Dir(filename))
		defer os.Remove(filename)
	case "application/zip":
		cmd = exec.Command("unzip", filename, "-d", filepath.Dir(filename))
		defer os.Remove(filename)
	default:
		return fmt.Errorf("invalid archive: %s", filename)
	}

	var out bytes.Buffer
	cmd.Stderr = &out

	err = cmd.Run()
	if err != nil {
		return errors.New(out.String())
	}

	if fileType == "application/gzip" {
		decompressedGzipFilename := strings.TrimSuffix(filename, filepath.Ext(filename))
		if filepath.Ext(filename) == ".tgz" {
			decompressedGzipFilename = decompressedGzipFilename + ".tar"
		}
		err = i.UnpackBlob(decompressedGzipFilename)
		if err != nil {
			if !strings.Contains(err.Error(), "invalid archive") {
				// not tested
				return err
			}
		}
	}

	return nil
}

func mimeType(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}

	buf := make([]byte, 512)
	_, err = file.Read(buf)
	if err != nil && err != io.EOF {
		// not tested
		return "", err
	}

	kind, err := filetype.Match(buf)
	if err != nil {
		// not tested
		return "", err
	}

	return kind.MIME.Value, nil

}

func min(x, y uint64) uint64 {
	if x < y {
		return x
	}
	return y
}

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

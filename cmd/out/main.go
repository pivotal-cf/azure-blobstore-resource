package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/Azure/azure-sdk-for-go/storage"
	"github.com/pivotal-cf/azure-blobstore-resource/api"
	"github.com/pivotal-cf/azure-blobstore-resource/azure"
)

const (
	BlobDefaultUploadBlockSize = 4 * 1024 * 1024 // 4 MB
	DefaultRetryTryTimeout     = time.Duration(0)
)

func main() {
	sourceDirectory := os.Args[1]

	var outRequest api.OutRequest
	err := json.NewDecoder(os.Stdin).Decode(&outRequest)
	if err != nil {
		log.Fatal("failed to decode: ", err)
	}

	baseURL := storage.DefaultBaseURL
	if outRequest.Source.BaseURL != "" {
		baseURL = outRequest.Source.BaseURL
	}

	azureClient := azure.NewClient(
		baseURL,
		outRequest.Source.StorageAccountName,
		outRequest.Source.StorageAccountKey,
		outRequest.Source.Container,
	)
	out := api.NewOut(azureClient)

	var blobName string
	var createSnapshot bool
	if outRequest.Source.VersionedFile != "" {
		blobName = outRequest.Source.VersionedFile
		createSnapshot = true
	} else if outRequest.Source.Regexp != "" {
		blobPath := filepath.Dir(outRequest.Source.Regexp)
		blobBaseName := filepath.Base(outRequest.Params.File)
		blobName = filepath.Join(blobPath, blobBaseName)
	}

	blockSize := BlobDefaultUploadBlockSize
	if outRequest.Params.BlockSize != nil {
		blockSize = *outRequest.Params.BlockSize
	}

	retryTryTimeout := DefaultRetryTryTimeout
	if outRequest.Params.Retry.TryTimeout != nil {
		retryTryTimeout = time.Duration(*outRequest.Params.Retry.TryTimeout)
	}

	path, snapshot, err := out.UploadFileToBlobstore(
		sourceDirectory,
		outRequest.Params.File,
		blobName,
		createSnapshot,
		blockSize,
		retryTryTimeout,
	)
	if err != nil {
		log.Fatal("failed to upload blob: ", err)
	}

	if createSnapshot {
		path = ""
	}

	versionsJSON, err := json.Marshal(api.Response{
		Version: api.ResponseVersion{
			Snapshot: snapshot,
			Path:     path,
		},
	})
	if err != nil {
		log.Fatal("failed to marshal output: ", err)
	}

	fmt.Println(string(versionsJSON))
}

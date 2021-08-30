package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"regexp"
	"time"

	"github.com/Azure/azure-sdk-for-go/storage"
	"github.com/cppforlife/go-semi-semantic/version"
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
		createSnapshot = false
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

	var ver version.Version
	if createSnapshot {
		path = ""
	} else {
		matcher, err := regexp.Compile(outRequest.Source.Regexp)
		if err != nil {
			log.Fatal("failed to compile source configuration regex: ", err)
		}

		matches := matcher.FindStringSubmatch(path)
		// No error if `len(matches) < 2` to preserve behaviour that the
		// resource doesn't error if the regex doesn't find a match in the
		// uploaded blob path
		if len(matches) >= 2 {
			var match string
			index, found := api.FindSubexpression(matcher.SubexpNames(), "version")
			if found {
				match = matches[index]
			} else {
				match = matches[1]
			}

			ver, err = version.NewVersionFromString(match)
			if err != nil {
				log.Fatal("failed to convert version from string: ", err)
			}
		}
	}

	versionsJSON, err := json.Marshal(api.Response{
		Version: api.ResponseVersion{
			Snapshot: snapshot,
			Path:     path,
			Version:  ver.AsString(),
		},
	})
	if err != nil {
		log.Fatal("failed to marshal output: ", err)
	}

	fmt.Println(string(versionsJSON))
}
